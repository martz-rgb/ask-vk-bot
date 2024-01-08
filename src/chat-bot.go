package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type ChatBot struct {
	cache      *Cache
	init_state StateNode

	timeout time.Duration

	api *VkApi
	db  *Db
}

func NewChatBot(init_state StateNode, timeout time.Duration, api *VkApi, db *Db) *ChatBot {
	return &ChatBot{
		cache:      NewCache(),
		init_state: init_state,
		timeout:    timeout,
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

	lp.MessageNew(bot.MessageNew)
	lp.MessageEvent(bot.MessageEvent)

	lp.RunWithContext(ctx)
}

func (bot *ChatBot) MessageNew(ctx context.Context, obj events.MessageNewObject) {
	user_id := obj.Message.FromID

	bot.api.MarkAsRead(user_id)

	existed := true
	chat, ok := bot.cache.Get(user_id)
	if !ok {
		existed = false
		chat = bot.cache.PutAndGet(user_id,
			NewChat(user_id,
				bot.init_state,
				bot.timeout,
				bot.cache.NotifyExpired))
	}
	defer bot.cache.Put(user_id, chat)

	chat.ResetTimer(bot.timeout, user_id, bot.cache.NotifyExpired)

	if !existed {
		chat.Init(bot.api, bot.db, user_id, false)
		return
	}
	next := chat.Do(bot.api, bot.db, NewMessageEvent, obj)
	if next != nil {
		chat.ChangeState(next)
		chat.Init(bot.api, bot.db, user_id, true)
	}
}

func (bot *ChatBot) MessageEvent(ctx context.Context, obj events.MessageEventObject) {
	user_id := obj.UserID

	bot.api.SendEventAnswer(obj.EventID, user_id, obj.PeerID)

	existed := true
	chat, ok := bot.cache.Get(user_id)
	if !ok {
		existed = false
		chat = bot.cache.PutAndGet(user_id,
			NewChat(user_id,
				bot.init_state,
				bot.timeout,
				bot.cache.NotifyExpired))
	}
	defer bot.cache.Put(user_id, chat)

	chat.ResetTimer(bot.timeout, user_id, bot.cache.NotifyExpired)

	if !existed {
		chat.Init(bot.api, bot.db, user_id, false)
		// and try to do next step
	}
	next := chat.Do(bot.api, bot.db, ChangeKeyboardEvent, obj)
	if next != nil {
		chat.ChangeState(next)
		chat.Init(bot.api, bot.db, user_id, true)
	}
}
