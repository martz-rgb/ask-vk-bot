package main

import (
	"sync"
	"time"
)

type Chat struct {
	user_id int
	state   StateNode

	in_use *sync.Mutex

	timer   *time.Timer
	expired time.Time
}

func (c *Chat) TimerFunc(expired chan int) func() {
	return func() {
		expired <- c.user_id
	}
}

func NewChat(user_id int, state StateNode, timeout time.Duration, expired chan int) *Chat {
	c := &Chat{
		user_id: user_id,
		in_use:  &sync.Mutex{},
		state:   state,
		expired: time.Now().Add(timeout),
	}
	c.timer = time.AfterFunc(timeout, c.TimerFunc(expired))

	return c
}

func (c *Chat) Entry(ask *Ask, vk *VK, silent bool) {
	c.state.Entry(ask, vk, c.user_id, silent)
}

func (c *Chat) Do(ask *Ask, vk *VK, event EventType, i interface{}) StateNode {
	return c.state.Do(ask, vk, event, i)
}

// reset timer and make new if timer was expired
func (c *Chat) ResetTimer(timeout time.Duration, expired chan int) {
	active := c.timer.Reset(timeout)
	if !active {
		c.timer = time.AfterFunc(timeout, c.TimerFunc(expired))
	}
}

func (c *Chat) ChangeState(next StateNode) *Chat {
	c.state = next
	return c
}
