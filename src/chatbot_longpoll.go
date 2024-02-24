package main

import (
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (bot *ChatBot) RunLongPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	lp, err := bot.controls.Vk.NewLongPoll()
	if err != nil {
		bot.log.Errorw("failed to run bot longpoll",
			"error", err,
			"id", bot.controls.Vk.ID())
		return
	}

	lp.MessageNew(bot.MessageNew)
	lp.MessageEvent(bot.MessageEvent)

	lp.RunWithContext(ctx)
}

func (bot *ChatBot) MessageNew(ctx context.Context, obj events.MessageNewObject) {
	user_id := obj.Message.FromID

	bot.controls.Vk.MarkAsRead(user_id)

	err := bot.Work(ctx, user_id, obj)
	if err != nil {
		bot.log.Errorw("error occured while new message",
			"user_id", user_id,
			"error", err)
	}
}

func (bot *ChatBot) MessageEvent(ctx context.Context, obj events.MessageEventObject) {
	user_id := obj.UserID

	bot.controls.Vk.SendEventAnswer(obj.EventID, user_id, obj.PeerID)

	err := bot.Work(ctx, user_id, obj)
	if err != nil {
		bot.log.Errorw("error occured while message event",
			"user_id", user_id,
			"error", err)
	}
}

func (bot *ChatBot) Work(ctx context.Context, user_id int, obj interface{}) error {
	chat, created := bot.TakeChat(user_id, &InitNode{})
	defer bot.ReturnChat(user_id)

	err := chat.Work(bot.controls, obj, created)
	if err != nil {
		return zaperr.Wrap(err, "",
			zap.String("state", chat.stack.Peek().ID()))
	}
	return nil
}
