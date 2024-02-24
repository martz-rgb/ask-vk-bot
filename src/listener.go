package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type Listener struct {
	ask *ask.Ask

	group *vk.VK
	admin *vk.VK

	log *zap.SugaredLogger
}

func NewListener(ask *ask.Ask, group *vk.VK, admin *vk.VK) *Listener {
	return &Listener{
		ask:   ask,
		group: group,
		admin: admin,
	}
}

func (l *Listener) RunLongPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	lp, err := l.group.NewLongPoll()
	if err != nil {
		zap.S().Errorw("failed to run listener longpoll",
			"error", err,
			"id", l.group.ID())
		return
	}

	lp.WallPostNew(l.WallPostNew)

	lp.RunWithContext(ctx)
}

func (l *Listener) WallPostNew(ctx context.Context, event events.WallPostNewObject) {
	zap.S().Info(event)

	//l.admin.WallPostNew(l.group_id, "got: "+event.Text, "", false, time.Now().Add(5*time.Minute).Unix())
}
