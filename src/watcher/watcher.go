package watcher

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"ask-bot/src/watcher/postponed"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Controls struct {
	Ask *ask.Ask

	Group *vk.VK
	Admin *vk.VK

	NotifyUser chan *vk.MessageParams
}

type Watcher struct {
	c *Controls
	p *postponed.Postponed

	log *zap.SugaredLogger
}

func New(controls *Controls, tick time.Duration, p *postponed.Postponed, log *zap.SugaredLogger) *Watcher {
	return &Watcher{
		c:   controls,
		p:   p,
		log: log,
	}
}

func (w *Watcher) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	go w.run(ctx, wg, w.c.CheckReservationsDeadline)
	go w.run(ctx, wg, w.c.CheckAlbums)
	go w.run(ctx, wg, w.c.CheckBoards)
	go w.run(ctx, wg, w.c.CheckPolls)
	go w.run(ctx, wg, w.updatePostponed)
}

func (w *Watcher) run(ctx context.Context, wg *sync.WaitGroup, exec func() error) {
	wg.Add(1)
	defer wg.Done()

	err := exec()
	if err != nil {
		w.log.Errorw("failed to exec",
			"error", err)
	}

	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ticker.C:
			err := exec()
			if err != nil {
				w.log.Errorw("failed to exec",
					"error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
