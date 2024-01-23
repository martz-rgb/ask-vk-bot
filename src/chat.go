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
	user_id     int
	state       StateNode
	reset_state StateNode

	in_use *sync.Mutex

	timer   *time.Timer
	expired time.Time
}

func NewChat(user_id int, state StateNode, reset_state StateNode, timeout time.Duration, expired chan int) *Chat {
	c := &Chat{
		user_id:     user_id,
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
		expired <- c.user_id
	}
}

// reset timer and make new if timer was expired
func (c *Chat) ResetTimer(timeout time.Duration, expired chan int) {
	active := c.timer.Reset(timeout)
	if !active {
		c.timer = time.AfterFunc(timeout, c.TimerFunc(expired))
	}
}

func (c *Chat) Work(ask *Ask, vk *VK, input interface{}, init int) error {
	switch event := input.(type) {
	case events.MessageNewObject:
		err := vk.MarkAsRead(c.user_id)
		if err != nil {
			c.Reset(vk)
			return err
		}

	case events.MessageEventObject:
		err := vk.SendEventAnswer(event.EventID, c.user_id, event.PeerID)
		if err != nil {
			c.Reset(vk)
			return err
		}
	}

	if init != NoInit {
		err := c.state.Entry(c.user_id, ask, vk, nil)
		if err != nil {
			c.Reset(vk)
			return err
		}

		if init == OnlyInit {
			return nil
		}
	}

	next, err := c.state.Do(c.user_id, ask, vk, input)
	if err != nil {
		c.Reset(vk)
		return err
	}

	if next != nil {
		c.state = next
		err := c.state.Entry(c.user_id, ask, vk, Params{"silent": true})
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
	vk.SendMessage(c.user_id, message, "", nil)
}
