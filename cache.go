package main

import "sync"

// to-do capacity of cache?
// or expiring users (more fair)
type Cache struct {
	sync.Mutex
	users map[int]StateNode
}

func NewCache() *Cache {
	return &Cache{
		users: make(map[int]StateNode),
	}
}

func (c *Cache) Load(key int) (StateNode, bool) {
	c.Lock()
	defer c.Unlock()

	user, ok := c.users[key]
	return user, ok
}

func (c *Cache) Store(key int, value StateNode) {
	c.Lock()
	defer c.Unlock()

	c.users[key] = value
}

func (c *Cache) LoadOrStore(key int, value StateNode) (actual StateNode, loaded bool) {
	c.Lock()
	defer c.Unlock()

	user, ok := c.users[key]
	if !ok {
		c.users[key] = value
		return value, false
	}

	return user, true
}

func (c *Cache) Delete(key int) {
	c.Lock()
	defer c.Unlock()

	delete(c.users, key)
}
