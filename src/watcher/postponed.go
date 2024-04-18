package watcher

import "ask-bot/src/watcher/postponed"

func (w *Watcher) updatePostponed() error {
	return w.p.Update(&postponed.Controls{
		Ask: w.c.Ask,
		Vk:  w.c.Admin,
	})
}
