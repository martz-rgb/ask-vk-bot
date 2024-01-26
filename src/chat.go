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
	state       StateNode
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
		state:       state,
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
		err := c.state.Entry(c.user, ask, vk, nil)
		if err != nil {
			c.Reset(vk)
			return err
		}

		if init == OnlyInit {
			return nil
		}
	}

	var next StateNode
	switch event := input.(type) {
	case events.MessageNewObject:
		next, err = c.state.NewMessage(c.user, ask, vk, event.Message.Text)
		if err != nil {
			c.Reset(vk)
			return err
		}

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(c.state, event.Payload)
		if err != nil {
			c.Reset(vk)
			return err
		}

		// skip
		if payload.Id != c.state.ID() {
			return nil
		}

		next, err = c.state.KeyboardEvent(c.user, ask, vk, payload)
		if err != nil {
			c.Reset(vk)
			return err
		}
	}

	if next != nil {
		c.state = next
		err := c.state.Entry(c.user, ask, vk, Params{"silent": true})
		if err != nil {
			c.Reset(vk)
			return err
		}
	}

	return nil
}

func (c *Chat) Reset(vk *VK) {
	c.state = c.reset_state

	message := "В ходе работы произошла ошибка. Пожалуйста, попробуйте еще раз попозже."
	vk.SendMessage(c.user.id, message, "", nil)
}
