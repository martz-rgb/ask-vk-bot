package chatbot

import (
	"ask-bot/src/ask"
	"ask-bot/src/chatbot/states"
	"ask-bot/src/postponed"
	"ask-bot/src/storage"
	"ask-bot/src/vk"
	"time"

	"go.uber.org/zap"
)

type Chatbot struct {
	cache       *storage.Storage[int, *Chat]
	reset_state states.State

	timeout time.Duration

	controls *states.Controls

	log *zap.SugaredLogger
}

func New(a *ask.Ask,
	v *vk.VK,
	p *postponed.Postponed,
	timeout time.Duration,
	log *zap.SugaredLogger) *Chatbot {
	bot := &Chatbot{
		cache:       storage.New[int, *Chat](),
		reset_state: &states.Init{},
		timeout:     timeout,
		controls: &states.Controls{
			Ask:       a,
			Vk:        v,
			Notify:    make(chan *vk.MessageParams),
			Postponed: p,
		},
		log: log,
	}

	go bot.ListenNotification()

	return bot
}

func (bot *Chatbot) TakeChat(user_id int, init states.State) (*Chat, bool) {
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

func (bot *Chatbot) ReturnChat(user_id int) {
	bot.cache.Return(user_id)
}

func (bot *Chatbot) ListenNotification() {
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

func (bot *Chatbot) NotifyChat(message *vk.MessageParams) error {
	chat, existed := bot.TakeChat(message.Id, &states.Init{
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
