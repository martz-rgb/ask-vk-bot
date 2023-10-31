package main

import (
	"context"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type ChatBot struct {
	cache       *Cache
	init_state  StateNode
	description StateMachine

	api *VkApi
	db  *Db
}

func NewChatBot(description StateMachine, init_state StateNode, api *VkApi, db *Db) *ChatBot {
	return &ChatBot{
		cache:       NewCache(),
		description: description,
		init_state:  init_state,
		api:         api,
		db:          db,
	}
}

func (b *ChatBot) RunLongPoll() {
	group, err := b.api.group.GroupsGetByID(api.Params{})
	if err != nil {
		panic(err)
	}

	fmt.Println("OK", group[0].ID)

	lp, err := longpoll.NewLongPoll(b.api.group, group[0].ID)
	if err != nil {
		panic(err)
	}

	lp.MessageNew(b.MessageEvent)
	lp.MessageEvent(b.KeyboardEvent)

	lp.Run()
}

func (b *ChatBot) MessageEvent(ctx context.Context, obj events.MessageNewObject) {
	user_id := obj.Message.FromID

	state, existed := b.cache.LoadOrStore(user_id, b.init_state)
	Entry, Do := b.description.GetNode(state)

	if !existed {
		Entry(b.api, user_id, false)
		return
	}

	next, change := Do(b.api, NewMessageEvent, obj)
	if !change {
		return
	}

	b.cache.Store(user_id, next)

	NextEntry, _ := b.description.GetNode(next)
	NextEntry(b.api, user_id, true)
}

func (b *ChatBot) KeyboardEvent(ctx context.Context, obj events.MessageEventObject) {
	b.api.SendEventAnswer(obj.EventID, obj.UserID, obj.PeerID)

	user_id := obj.UserID

	state, existed := b.cache.LoadOrStore(user_id, b.init_state)
	Entry, Do := b.description.GetNode(state)

	if !existed {
		Entry(b.api, user_id, false)
		return
	}

	next, change := Do(b.api, ChangeKeyboardEvent, obj)
	if !change {
		return
	}

	b.cache.Store(user_id, next)

	NextEntry, _ := b.description.GetNode(next)
	NextEntry(b.api, user_id, true)
}
