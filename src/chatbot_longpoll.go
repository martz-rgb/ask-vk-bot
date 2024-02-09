package main

import (
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

func (bot *ChatBot) RunLongPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	lp, err := longpoll.NewLongPoll(bot.vk.api, bot.vk.id)
	if err != nil {
		bot.log.Errorw("failed to run bot longpoll",
			"error", err,
			"id", bot.vk.id)
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

	init := NoInit
	if !existed {
		init = OnlyInit
	}

	err := chat.Work(bot.ask, bot.vk, obj, init)
	if err != nil {
		bot.log.Errorw("error occured while new message",
			"user_id", user_id,
			"state", chat.stack.Peek().ID(),
			"error", err)
	}
}

func (bot *ChatBot) MessageEvent(ctx context.Context, obj events.MessageEventObject) {
	user_id := obj.UserID

	bot.vk.SendEventAnswer(obj.EventID, user_id, obj.PeerID)

	chat, existed := bot.GetChat(user_id)
	defer bot.PutChat(user_id, chat)

	chat.ResetTimer(bot.timeout, bot.cache.NotifyExpired)

	init := NoInit
	if !existed {
		init = InitAndTry
	}

	err := chat.Work(bot.ask, bot.vk, obj, init)
	if err != nil {
		bot.log.Errorw("error occured while message event",
			"user_id", user_id,
			"state", chat.stack.Peek().ID(),
			"error", err)
	}
}
