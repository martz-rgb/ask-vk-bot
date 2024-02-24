package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"time"

	"go.uber.org/zap"
)

type Controls struct {
	Ask    *ask.Ask
	Vk     *vk.VK
	Notify chan *vk.MessageParams
}

type ChatBot struct {
	cache       *Cache[int, *Chat]
	reset_state StateNode

	timeout time.Duration

	controls *Controls

	log *zap.SugaredLogger
}

func NewChatBot(a *ask.Ask, v *vk.VK, reset_state StateNode, timeout time.Duration, log *zap.SugaredLogger) *ChatBot {
	bot := &ChatBot{
		cache:       NewCache[int, *Chat](),
		reset_state: reset_state,
		timeout:     timeout,
		controls: &Controls{
			Ask:    a,
			Vk:     v,
			Notify: make(chan *vk.MessageParams),
		},
		log: log,
	}

	go bot.ListenNotification()

	return bot
}

func (bot *ChatBot) TakeChat(user_id int, init StateNode) (*Chat, bool) {
	chat, existed := bot.cache.CreateIfNotExistedAndTake(user_id,
		NewChat(user_id,
			init,
			bot.reset_state,
			bot.timeout,
			bot.cache.NotifyExpired,
			bot.controls))

	chat.ResetTimer(bot.timeout, bot.cache.NotifyExpired, bot.controls)
	return chat, existed
}

func (bot *ChatBot) ReturnChat(user_id int) {
	bot.cache.Return(user_id)
}

func (bot *ChatBot) ListenNotification() {
	for {
		message := <-bot.controls.Notify
		err := bot.NotifyChat(message)
		if err != nil {
			bot.log.Errorw("error occured while try to notify",
				"message", message,
				"error", err)
		}
	}
}

func (bot *ChatBot) NotifyChat(message *vk.MessageParams) error {
	chat, existed := bot.TakeChat(message.Id, &InitNode{
		Silent: true,
	})
	defer bot.ReturnChat(message.Id)

	if !existed {
		chat.Work(bot.controls, nil, true)
	}

	err := chat.Notify(bot.controls, message)
	if err != nil {
		return err
	}

	return nil
}
