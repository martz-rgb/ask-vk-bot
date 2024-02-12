package main

import (
	"sync"
	"time"

	"github.com/SevereCloud/vksdk/v2/events"
)

const (
	NoInit = iota
	OnlyInit
	InitAndTry
)

type Chat struct {
	user        *User
	stack       *StateStack
	reset_state StateNode

	in_use *sync.Mutex

	timer   *time.Timer
	expired time.Time
}

func NewChat(user_id int, state StateNode, reset_state StateNode, timeout time.Duration, expired chan int) *Chat {
	c := &Chat{
		user: &User{
			id: user_id,
		},
		in_use:      &sync.Mutex{},
		stack:       &StateStack{state},
		reset_state: reset_state,
		expired:     time.Now().Add(timeout),
	}
	c.timer = time.AfterFunc(timeout, c.TimerFunc(expired))

	return c
}

func (c *Chat) TimerFunc(expired chan int) func() {
	return func() {
		expired <- c.user.id
	}
}

// reset timer and make new if timer was expired
func (c *Chat) ResetTimer(timeout time.Duration, expired chan int) {
	active := c.timer.Reset(timeout)
	if !active {
		c.timer = time.AfterFunc(timeout, c.TimerFunc(expired))
	}
}

const (
	NewMessage = iota
	KeyboardEvent
)

func (c *Chat) Work(ask *Ask, vk *VK, input interface{}, init int) (err error) {
	if init != NoInit {
		err := c.stack.Peek().Entry(c.user, ask, vk)
		if err != nil {
			c.Reset(vk)
			return err
		}

		if init == OnlyInit {
			return nil
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

		next, back, err = c.stack.Peek().NewMessage(c.user, ask, vk, message)
		if err != nil {
			c.Reset(vk)
			return err
		}

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(c.stack.Peek(), event.Payload)
		if err != nil {
			c.Reset(vk)
			return err
		}

		// skip
		if payload.Id != c.stack.Peek().ID() {
			return nil
		}

		next, back, err = c.stack.Peek().KeyboardEvent(c.user, ask, vk, payload)
		if err != nil {
			c.Reset(vk)
			return err
		}
	}

	if next != nil {
		c.stack.Push(next)
		err := c.stack.Peek().Entry(c.user, ask, vk)
		if err != nil {
			c.Reset(vk)
			return err
		}
	} else if back {
		for back {
			prev := c.stack.Pop()
			back, err = c.stack.Peek().Back(c.user, ask, vk, prev)
			if err != nil {
				c.Reset(vk)
				return err
			}
		}
	}

	return nil
}

func (c *Chat) Reset(vk *VK) {
	c.stack = &StateStack{c.reset_state}

	message := "В ходе работы произошла ошибка. Пожалуйста, попробуйте еще раз попозже."
	vk.SendMessage(c.user.id, message, "", nil)
}
