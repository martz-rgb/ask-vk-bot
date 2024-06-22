package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/posts"
	"ask-bot/src/vk"
	"sync"
	"time"
)

type Dictionary map[posts.Kind][]posts.Post

type Controls struct {
	Ask *ask.Ask
	Vk  *vk.VK
}

type Postponed struct {
	mu    *sync.Mutex
	posts Dictionary
	busy  []time.Time
}

func New() *Postponed {
	return &Postponed{
		&sync.Mutex{},
		nil,
		nil,
	}
}

func (p *Postponed) Update(c *Controls) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	db, err := NewDBInfo(c)
	if err != nil {
		return err
	}

	vk, err := NewVKInfo(c)
	if err != nil {
		return err
	}

	return p.update(c, db, vk)
}
