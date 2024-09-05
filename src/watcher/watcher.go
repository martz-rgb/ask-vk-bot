package watcher

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"ask-bot/src/watcher/events"
	"ask-bot/src/watcher/postponed"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TO-DO how to divide group/admin vk api & postponed
type Controls struct {
	Ask *ask.Ask

	Group *vk.VK
	Admin *vk.VK

	Postponed *postponed.Postponed

	NotifyUser  chan *vk.MessageParams
	NotifyEvent chan events.Event
}

func (c *Controls) PostponedControls() *postponed.Controls {
	return &postponed.Controls{
		Vk:  c.Admin,
		Ask: c.Ask,
	}
}

var buf = 1

var notifications = struct {
	Album chan bool
	Board chan bool
}{
	make(chan bool, buf),
	make(chan bool, buf),
}

type Watcher struct {
	c *Controls

	log *zap.SugaredLogger
}

func New(controls *Controls, tick time.Duration, log *zap.SugaredLogger) *Watcher {
	return &Watcher{
		c:   controls,
		log: log,
	}
}

func (w *Watcher) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	go w.listenEvents(ctx, wg)

	go w.runWithNotify(ctx, wg, w.c.CheckAlbums, notifications.Album)
	go w.runWithNotify(ctx, wg, w.c.CheckBoards, notifications.Board)

	go w.run(ctx, wg, w.c.CheckReservationsDeadline)

	go w.run(ctx, wg, w.c.UpdatePostponed)
	go w.run(ctx, wg, w.c.DeleteInvalidPostponed)

	go w.run(ctx, wg, w.c.CheckPendingPolls)
	go w.run(ctx, wg, w.c.CheckOngoingPolls)
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

func (w *Watcher) runWithNotify(ctx context.Context, wg *sync.WaitGroup, exec func() error, notify chan bool) {
	wg.Add(1)
	defer wg.Done()

	err := exec()
	if err != nil {
		w.log.Errorw("failed to exec",
			"error", err)
	}

	ticker := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-ticker.C:
			err := exec()
			if err != nil {
				w.log.Errorw("failed to exec on ticker",
					"error", err)
			}
		case <-notify:
			err := exec()
			if err != nil {
				w.log.Errorw("failed to exec on notify",
					"error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
