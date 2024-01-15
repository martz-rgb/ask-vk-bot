package main

import (
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"go.uber.org/zap"
)

func (bot *ChatBot) RunLongPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	lp, err := longpoll.NewLongPoll(bot.vk.api, bot.group_id)
	if err != nil {
		zap.S().Errorw("failed to run bot longpoll",
			"error", err,
			"id", bot.group_id)
		return
	}

	lp.MessageNew(bot.MessageNew)
	lp.MessageEvent(bot.MessageEvent)

	lp.RunWithContext(ctx)
}

func (bot *ChatBot) MessageNew(ctx context.Context, obj events.MessageNewObject) {
	user_id := obj.Message.FromID

	bot.vk.MarkAsRead(user_id)

	chat, existed := bot.GetChat(user_id)
	defer bot.PutChat(user_id, chat)

	chat.ResetTimer(bot.timeout, bot.cache.NotifyExpired)

	if !existed {
		chat.Entry(bot.ask, bot.vk, false)
		return
	}
	next := chat.Do(bot.ask, bot.vk, NewMessageEvent, obj)
	if next != nil {
		chat.ChangeState(next)
		chat.Entry(bot.ask, bot.vk, true)
	}
}

func (bot *ChatBot) MessageEvent(ctx context.Context, obj events.MessageEventObject) {
	user_id := obj.UserID

	bot.vk.SendEventAnswer(obj.EventID, user_id, obj.PeerID)

	chat, existed := bot.GetChat(user_id)
	defer bot.PutChat(user_id, chat)

	chat.ResetTimer(bot.timeout, bot.cache.NotifyExpired)

	if !existed {
		chat.Entry(bot.ask, bot.vk, false)
		// and try to do next step
	}
	next := chat.Do(bot.ask, bot.vk, ChangeKeyboardEvent, obj)
	if next != nil {
		chat.ChangeState(next)
		chat.Entry(bot.ask, bot.vk, true)
	}
}
