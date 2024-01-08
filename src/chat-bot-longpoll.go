package main

import (
	"context"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

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

	chat, existed := bot.GetChat(user_id)
	defer bot.PutChat(user_id, chat)

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

	chat, existed := bot.GetChat(user_id)
	defer bot.PutChat(user_id, chat)

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
