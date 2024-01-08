package main

import (
	"sync"
	"time"
)

// vk longpoll is sequential, but deleting with timer is not
type Cache struct {
	*sync.Mutex

	NotifyExpired chan int
	chats         map[int]*Chat
}

func NewCache() *Cache {
	cache := &Cache{
		&sync.Mutex{},
		make(chan int),
		make(map[int]*Chat),
	}

	go cache.ListenExpired()

	return cache
}

func (c *Cache) Get(key int) (actual *Chat, ok bool) {
	c.Lock()
	defer c.Unlock()

	chat, ok := c.chats[key]
	if !ok {
		return nil, false
	}

	chat.in_use.Lock()
	return chat, true
}

func (c *Cache) Put(key int, value *Chat) {
	c.Lock()
	defer c.Unlock()

	_, ok := c.chats[key]
	c.chats[key] = value

	if ok {
		value.in_use.Unlock()
		return
	}
}

func (c *Cache) PutAndGet(key int, value *Chat) *Chat {
	c.Lock()
	defer c.Unlock()

	c.chats[key] = value
	value.in_use.Lock()

	return value
}

func (c *Cache) ListenExpired() {
	for {
		key := <-c.NotifyExpired

		c.Lock()
		value, ok := c.chats[key]
		if ok {
			value.in_use.Lock()
			if time.Now().After(value.expired) {
				delete(c.chats, key)
			}
			value.in_use.Unlock()
		}
		c.Unlock()
	}
}
