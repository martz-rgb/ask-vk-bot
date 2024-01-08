package main

import (
	"time"
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
