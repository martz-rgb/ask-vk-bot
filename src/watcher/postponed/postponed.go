package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"sync"
)

type Controls struct {
	Ask *ask.Ask
	Vk  *vk.VK
}

type Postponed struct {
	mu    *sync.Mutex
	cache *Cache
}

func New() *Postponed {
	return &Postponed{
		&sync.Mutex{},
		&Cache{},
	}
}

func (p *Postponed) Update(c *Controls) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.cache.internalUpdate(c)
	if err != nil {
		return err
	}

	db, err := NewDBInfo(c)
	if err != nil {
		return err
	}

	vk, err := NewVKInfo(c, p.cache)
	if err != nil {
		return err
	}

	return p.cache.update(c, db, vk)
}
