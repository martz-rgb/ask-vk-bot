package main

import (
	"context"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type ChatBot struct {
	cache      *Cache
	init_state StateNode

	api *VkApi
	db  *Db
}

func NewChatBot(init_state StateNode, api *VkApi, db *Db) *ChatBot {
	return &ChatBot{
		cache:      NewCache(),
		init_state: init_state,
		api:        api,
		db:         db,
	}
}

func (bot *ChatBot) RunLongPoll(ctx context.Context) {
	group, err := bot.api.group.GroupsGetByID(api.Params{})
	if err != nil {
		panic(err)
	}

	fmt.Println("OK", group[0].ID)

	lp, err := longpoll.NewLongPoll(bot.api.group, group[0].ID)
	if err != nil {
		panic(err)
	}

	lp.MessageNew(bot.MessageEvent)
	lp.MessageEvent(bot.KeyboardEvent)

	lp.RunWithContext(ctx)
}

func (bot *ChatBot) MessageEvent(ctx context.Context, obj events.MessageNewObject) {
	user_id := obj.Message.FromID

	bot.api.MarkAsRead(user_id)

	state, existed := bot.cache.LoadOrStore(user_id, bot.init_state)

	if !existed {
		state.Init(bot.api, bot.db, user_id, false)
		return
	}

	next := state.Do(bot.api, bot.db, NewMessageEvent, obj)
	if next != nil {
		bot.cache.Store(user_id, next)
		next.Init(bot.api, bot.db, user_id, true)
	}
}

func (bot *ChatBot) KeyboardEvent(ctx context.Context, obj events.MessageEventObject) {
	bot.api.SendEventAnswer(obj.EventID, obj.UserID, obj.PeerID)

	user_id := obj.UserID

	state, existed := bot.cache.LoadOrStore(user_id, bot.init_state)

	if !existed {
		state.Init(bot.api, bot.db, user_id, false)
		// and try to do next step
	}

	next := state.Do(bot.api, bot.db, ChangeKeyboardEvent, obj)
	if next != nil {
		bot.cache.Store(user_id, next)
		next.Init(bot.api, bot.db, user_id, true)
	}
}
