package main

import (
	"sync"
	"time"
)

type Chat struct {
	in_use *sync.Mutex

	state StateNode

	timer   *time.Timer
	expired time.Time
}

func ChatTimerFunc(user_id int, expired chan int) func() {
	return func() {
		expired <- user_id
	}
}

func NewChat(user_id int, state StateNode, timeout time.Duration, expired chan int) *Chat {
	return &Chat{
		in_use:  &sync.Mutex{},
		state:   state,
		timer:   time.AfterFunc(timeout, ChatTimerFunc(user_id, expired)),
		expired: time.Now().Add(timeout),
	}
}

func (chat *Chat) Init(a *VkApi, db *Db, user_id int, silent bool) {
	chat.state.Init(a, db, user_id, silent)
}

func (chat *Chat) Do(a *VkApi, db *Db, event EventType, i interface{}) StateNode {
	return chat.state.Do(a, db, event, i)
}

// reset timer and make new if timer was expired
func (chat *Chat) ResetTimer(timeout time.Duration, user_id int, expired chan int) {
	active := chat.timer.Reset(timeout)
	if !active {
		chat.timer = time.AfterFunc(timeout, ChatTimerFunc(user_id, expired))
	}
}

func (chat *Chat) ChangeState(next StateNode) *Chat {
	chat.state = next
	return chat
}
