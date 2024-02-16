package main

import (
	"time"

	"go.uber.org/zap"
)

type Controls struct {
	Ask    *Ask
	Vk     *VK
	Notify chan *MessageParams
}

type ChatBot struct {
	cache       *Cache[int, *Chat]
	reset_state StateNode

	timeout time.Duration

	controls *Controls

	log *zap.SugaredLogger
}

func NewChatBot(ask *Ask, reset_state StateNode, timeout time.Duration, vk *VK, log *zap.SugaredLogger) *ChatBot {
	bot := &ChatBot{
		cache:       NewCache[int, *Chat](),
		reset_state: reset_state,
		timeout:     timeout,
		controls: &Controls{
			Ask:    ask,
			Vk:     vk,
			Notify: make(chan *MessageParams),
		},
		log: log,
	}

	go bot.ListenNotification()

	return bot
}

func (bot *ChatBot) TakeChat(user_id int, init StateNode) (*Chat, bool) {
	return bot.cache.CreateIfNotExistedAndTake(user_id,
		NewChat(user_id,
			init,
			bot.reset_state,
			bot.timeout,
			bot.cache.NotifyExpired,
			bot.controls))
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

func (bot *ChatBot) NotifyChat(message *MessageParams) error {
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
