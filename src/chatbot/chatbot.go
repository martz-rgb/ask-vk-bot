package chatbot

import (
	"ask-bot/src/chatbot/states"
	"ask-bot/src/datatypes/storage"
	"ask-bot/src/vk"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Chatbot struct {
	storage     *storage.Storage[int, *Chat]
	reset_state states.State

	controls *states.Controls

	timeout time.Duration

	log *zap.SugaredLogger
}

type Controls states.Controls

func New(c *Controls,
	timeout time.Duration,
	log *zap.SugaredLogger) *Chatbot {
	controls := (*states.Controls)(c)

	bot := &Chatbot{
		storage:     storage.New[int, *Chat](),
		reset_state: &states.Init{},
		timeout:     timeout,
		controls:    controls,
		log:         log,
	}

	return bot
}

func (bot *Chatbot) TakeChat(user_id int, init states.State) (*Chat, bool) {
	chat, existed := bot.storage.CreateIfNotExistedAndTake(user_id,
		NewChat(user_id,
			init,
			bot.reset_state,
			bot.timeout,
			bot.storage.NotifyExpired,
			bot.controls))

	chat.ResetTimer(bot.timeout, bot.storage.NotifyExpired, bot.controls)
	return chat, existed
}

func (bot *Chatbot) ReturnChat(user_id int) {
	bot.storage.Return(user_id)
}

func (bot *Chatbot) ListenNotification(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case message := <-bot.controls.Notify:
			err := bot.NotifyChat(message)
			if err != nil {
				bot.log.Errorw("error occured while try to notify",
					"message", message,
					"error", err)
			}
		case <-ctx.Done():
			wg.Done()
			return
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
