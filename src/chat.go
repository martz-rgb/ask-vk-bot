package main

import (
	"sync"
	"time"
)

// state cache probably
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
	c.state.Entry(c.user_id, ask, vk, silent)
}

func (c *Chat) Do(ask *Ask, vk *VK, input interface{}) StateNode {
	return c.state.Do(c.user_id, ask, vk, input)
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
