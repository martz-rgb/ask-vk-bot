package chatbot

import (
	"ask-bot/src/chatbot/states"
	"ask-bot/src/stack"
	"ask-bot/src/vk"
	"time"

	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type Chat struct {
	user        *states.User
	stack       *stack.Stack[states.State]
	reset_state states.State
	query       []*vk.MessageParams

	timer *time.Timer
}

func NewChat(user_id int, state states.State, reset_state states.State, timeout time.Duration, expired chan int, controls *states.Controls) *Chat {
	return &Chat{
		user:        &states.User{Id: user_id},
		stack:       stack.New[states.State](state),
		reset_state: reset_state,
	}
}

func (c *Chat) TimerFunc(expired chan int, controls *states.Controls) func() {
	return func() {
		c.Finish(controls)

		expired <- c.user.Id
	}
}

// reset timer and make new if timer was expired
func (c *Chat) ResetTimer(timeout time.Duration, expired chan int, controls *states.Controls) {
	if c.timer == nil {
		c.timer = time.AfterFunc(timeout, c.TimerFunc(expired, controls))
		return
	}

	c.timer.Stop()
	c.timer.Reset(timeout)
}

func (c *Chat) tryNotify(controls *states.Controls) error {
	if c.stack.Len() <= 1 {
		for len(c.query) > 0 {
			notification := c.query[0]

			_, err := controls.Vk.SendMessage(c.user.Id,
				notification.Text,
				"",
				notification.Params)
			if err != nil {
				return err
			}

			c.query = c.query[1:]
		}
	}

	return nil
}

func (c *Chat) Notify(controls *states.Controls, message *vk.MessageParams) error {
	c.query = append(c.query, message)

	return c.tryNotify(controls)
}

func (c *Chat) Work(controls *states.Controls, input interface{}, init bool) error {
	err := c.work(controls, input, init)
	if err != nil {
		return err
	}

	return c.tryNotify(controls)
}

func (c *Chat) work(controls *states.Controls, input interface{}, existed bool) (err error) {
	if !existed {
		err := c.stack.Peek().Entry(c.user, controls)
		if err != nil {
			c.Reset(controls)
			return err
		}
	}

	var action *states.Action

	switch event := input.(type) {
	case events.MessageNewObject:
		message := &vk.Message{
			ID:          event.Message.ID,
			Text:        event.Message.Text,
			Attachments: event.Message.Attachments,
		}

		action, err = c.stack.Peek().NewMessage(c.user, controls, message)
		if err != nil {
			c.Reset(controls)
			return err
		}

	case events.MessageEventObject:
		payload, err := vk.UnmarshalPayload(event.Payload)
		if err != nil {
			c.Reset(controls)
			return err
		}

		// skip
		if payload.Id != c.stack.Peek().ID() {
			zap.S().Infow("wrong payload id",
				"payload", payload,
				"state", c.stack.Peek().ID())
			return nil
		}

		action, err = c.stack.Peek().KeyboardEvent(c.user, controls, payload)
		if err != nil {
			c.Reset(controls)
			return err
		}
	}

	switch action.Kind() {
	case states.Next:
		err = c.next(controls, action.Next())
	case states.Exit:
		err = c.exit(controls, action.Exit())
	}

	if err != nil {
		c.Reset(controls)
		return err
	}
	return nil
}

func (c *Chat) next(controls *states.Controls, next states.State) error {
	c.stack.Push(next)
	err := c.stack.Peek().Entry(c.user, controls)
	if err != nil {
		return err
	}

	return nil
}

func (c *Chat) exit(controls *states.Controls, info *states.ExitInfo) error {
	c.stack.Pop()

	action, err := c.stack.Peek().Back(c.user, controls, info)
	if err != nil {
		return err
	}

	for action != nil {
		switch action.Kind() {
		case states.Next:
			err := c.next(controls, action.Next())
			if err != nil {
				return err
			}
			return nil

		case states.Exit:
			c.stack.Pop()

			action, err = c.stack.Peek().Back(c.user, controls, action.Exit())
			if err != nil {
				return err
			}

		default:
			return nil
		}
	}

	return nil
}

func (c *Chat) Finish(controls *states.Controls) {
	for len(c.query) > 0 {
		notification := c.query[0]

		_, err := controls.Vk.SendMessage(c.user.Id,
			notification.Text,
			"",
			notification.Params)
		if err != nil {
			zap.S().Errorw("failed to send notification after timeout",
				"user", c.user,
				"notification", notification,
				"error", err)
			return
		}

		c.query = c.query[1:]
	}
}

func (c *Chat) Reset(controls *states.Controls) {
	c.stack = stack.New[states.State](c.reset_state)

	message := "В ходе работы произошла ошибка. Пожалуйста, попробуйте еще раз попозже."
	controls.Vk.SendMessage(c.user.Id, message, "", nil)

	c.stack.Peek().Back(c.user, controls, nil)
}
