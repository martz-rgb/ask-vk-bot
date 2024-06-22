package watcher

import (
	"ask-bot/src/watcher/events"
	"context"
	"sync"
)

type Action func()

var eventActions = map[events.Event]Action{
	events.NewRole: func() {
		select {
		case notifications.Album <- true:
		default:
		}

		select {
		case notifications.Board <- true:
		default:
		}
	},
}

func (w *Watcher) listenEvents(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	select {
	case event := <-w.c.NotifyEvent:
		eventActions[event]()
	case <-ctx.Done():
		return
	}

}
