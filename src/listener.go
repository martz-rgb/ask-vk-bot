package main

import (
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"go.uber.org/zap"
)

type Listener struct {
	ask *Ask

	group *VK
	admin *VK
}

func NewListener(ask *Ask, group *VK, admin *VK) *Listener {
	return &Listener{
		ask:   ask,
		group: group,
		admin: admin,
	}
}

func (l *Listener) RunLongPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	lp, err := longpoll.NewLongPoll(l.group.api, l.group.id)
	if err != nil {
		zap.S().Errorw("failed to run listener longpoll",
			"error", err,
			"id", l.group.id)
		return
	}

	lp.WallPostNew(l.WallPostNew)

	lp.RunWithContext(ctx)
}

func (l *Listener) WallPostNew(ctx context.Context, event events.WallPostNewObject) {
	zap.S().Info(event)

	//l.admin.WallPostNew(l.group_id, "got: "+event.Text, "", false, time.Now().Add(5*time.Minute).Unix())
}
