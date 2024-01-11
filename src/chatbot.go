package main

import (
	"time"
)

type ChatBot struct {
	cache      *Cache
	init_state StateNode

	timeout time.Duration

	ask *Ask
	vk  *VK
}

func NewChatBot(ask *Ask, init_state StateNode, timeout time.Duration, vk *VK) *ChatBot {
	return &ChatBot{
		cache:      NewCache(),
		init_state: init_state,
		timeout:    timeout,
		ask:        ask,
		vk:         vk,
	}
}

func (bot *ChatBot) GetChat(user_id int) (*Chat, bool) {
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

	return chat, existed
}

func (bot *ChatBot) PutChat(user_id int, chat *Chat) {
	bot.cache.Put(user_id, chat)
}
