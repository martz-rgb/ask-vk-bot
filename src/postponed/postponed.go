package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Controls struct {
	Vk  *vk.VK
	Ask *ask.Ask
}

type Postponed struct {
	mu     *sync.Mutex
	notify chan bool

	c    *Controls
	tick time.Duration

	cache *Cache

	log *zap.SugaredLogger
}

func New(c *Controls, tick time.Duration, log *zap.SugaredLogger) (*Postponed, chan bool) {
	p := &Postponed{
		&sync.Mutex{},
		make(chan bool),
		c,
		tick,
		nil,
		log,
	}
	return p, p.notify
}

func (p *Postponed) Run(ctx context.Context, wg *sync.WaitGroup) {
	p.cache = NewCache(p.c)
	err := p.cache.internalUpdate(p.c)
	if err != nil {
		p.log.Errorw("failed to update cache",
			"error", err)
	}

	start := time.Now()
	err = p.update()
	elapsed := time.Since(start)
	if err != nil {
		p.log.Errorw("failed to update postponed on ticker",
			"error", err)
	}

	p.log.Debugw("update works for",
		"elapsed", elapsed)

	ticker := time.NewTicker(p.tick)

	for {
		select {
		case <-ticker.C:
			start := time.Now()

			err := p.update()

			elapsed := time.Since(start)
			if err != nil {
				p.log.Errorw("failed to update postponed on ticker",
					"error", err)
			}

			p.log.Debugw("update works for",
				"elapsed", elapsed)

		case <-p.notify:
			err := p.update()
			if err != nil {
				p.log.Errorw("failed to update postponed on notify",
					"error", err)
			}
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func (p *Postponed) update() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.cache.internalUpdate(p.c)
	if err != nil {
		return err
	}

	db, err := NewDBInfo(p.c)
	if err != nil {
		return err
	}

	vk, err := NewVKInfo(p.c, p.cache)
	if err != nil {
		return err
	}

	return p.cache.update(p.c, db, vk)
}
