package listener

import (
	"ask-bot/src/ask"
	"ask-bot/src/postponed"
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

	postponed *postponed.Postponed
	notify    chan bool

	log *zap.SugaredLogger
}

func New(
	ask *ask.Ask,
	group *vk.VK,
	admin *vk.VK,
	p *postponed.Postponed,
	notify chan bool) *Listener {
	return &Listener{
		ask:       ask,
		group:     group,
		admin:     admin,
		postponed: p,
		notify:    notify,
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
