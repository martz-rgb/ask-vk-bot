package main

import (
	"sync"
	"time"
)

// vk longpoll is sequential, but deleting with timer is not
type Cache struct {
	mu *sync.Mutex

	NotifyExpired chan int
	chats         map[int]*Chat
}

func NewCache() *Cache {
	cache := &Cache{
		mu:            &sync.Mutex{},
		NotifyExpired: make(chan int),
		chats:         make(map[int]*Chat),
	}

	go cache.ListenExpired()

	return cache
}

func (c *Cache) Get(key int) (actual *Chat, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	chat, ok := c.chats[key]
	if !ok {
		return nil, false
	}

	chat.in_use.Lock()
	return chat, true
}

func (c *Cache) Put(key int, value *Chat) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.chats[key]
	c.chats[key] = value

	if ok {
		value.in_use.Unlock()
		return
	}
}

func (c *Cache) PutAndGet(key int, value *Chat) *Chat {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chats[key] = value
	value.in_use.Lock()

	return value
}

func (c *Cache) ListenExpired() {
	for {
		key := <-c.NotifyExpired

		c.mu.Lock()
		value, ok := c.chats[key]
		if ok {
			value.in_use.Lock()
			if time.Now().After(value.expired) {
				delete(c.chats, key)
			}
			value.in_use.Unlock()
		}
		c.mu.Unlock()
	}
}
