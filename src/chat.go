package main

import (
	"time"

	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type Chat struct {
	user        *User
	stack       *StateStack
	reset_state StateNode
	query       []*MessageParams

	timer *time.Timer
}

func NewChat(user_id int, state StateNode, reset_state StateNode, timeout time.Duration, expired chan int, controls *Controls) *Chat {
	return &Chat{
		user: &User{
			id: user_id,
		},
		stack:       &StateStack{state},
		reset_state: reset_state,
	}
}

func (c *Chat) TimerFunc(expired chan int, controls *Controls) func() {
	return func() {
		c.Exit(controls)

		expired <- c.user.id
	}
}

// reset timer and make new if timer was expired
func (c *Chat) ResetTimer(timeout time.Duration, expired chan int, controls *Controls) {
	if c.timer == nil {
		c.timer = time.AfterFunc(timeout, c.TimerFunc(expired, controls))
		return
	}

	c.timer.Stop()
	c.timer.Reset(timeout)
}

func (c *Chat) tryNotify(controls *Controls) error {
	if c.stack.Len() <= 1 {
		for len(c.query) > 0 {
			notification := c.query[0]

			_, err := controls.Vk.SendMessage(c.user.id,
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

func (c *Chat) Notify(controls *Controls, message *MessageParams) error {
	c.query = append(c.query, message)

	return c.tryNotify(controls)
}

func (c *Chat) Work(controls *Controls, input interface{}, init bool) error {
	err := c.work(controls, input, init)
	if err != nil {
		return err
	}

	return c.tryNotify(controls)
}

func (c *Chat) work(controls *Controls, input interface{}, init bool) (err error) {
	if init {
		err := c.stack.Peek().Entry(c.user, controls)
		if err != nil {
			c.Reset(controls)
			return err
		}
	}

	var next StateNode
	var back bool

	switch event := input.(type) {
	case events.MessageNewObject:
		message := &Message{
			ID:          event.Message.ID,
			Text:        event.Message.Text,
			Attachments: event.Message.Attachments,
		}

		next, back, err = c.stack.Peek().NewMessage(c.user, controls, message)
		if err != nil {
			c.Reset(controls)
			return err
		}

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(c.stack.Peek(), event.Payload)
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

		next, back, err = c.stack.Peek().KeyboardEvent(c.user, controls, payload)
		if err != nil {
			c.Reset(controls)
			return err
		}
	}

	if next != nil {
		c.stack.Push(next)
		err := c.stack.Peek().Entry(c.user, controls)
		if err != nil {
			c.Reset(controls)
			return err
		}
	} else if back {
		for back {
			prev := c.stack.Pop()
			back, err = c.stack.Peek().Back(c.user, controls, prev)
			if err != nil {
				c.Reset(controls)
				return err
			}
		}
	}

	return nil
}

func (c *Chat) Exit(controls *Controls) {
	for len(c.query) > 0 {
		notification := c.query[0]

		_, err := controls.Vk.SendMessage(c.user.id,
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

func (c *Chat) Reset(controls *Controls) {
	c.stack = &StateStack{c.reset_state}

	message := "В ходе работы произошла ошибка. Пожалуйста, попробуйте еще раз попозже."
	controls.Vk.SendMessage(c.user.id, message, "", nil)

	c.stack.Peek().Back(c.user, controls, nil)
}
