package main

import (
	"sync"
	"sync/atomic"
)

type Record[T interface{}] struct {
	in_use  *sync.Mutex
	waiting *atomic.Int32

	value T
}

type Cache[K comparable, T interface{}] struct {
	mu *sync.Mutex

	NotifyExpired chan K
	records       map[K]*Record[T]
}

func NewCache[K comparable, T interface{}]() *Cache[K, T] {
	cache := &Cache[K, T]{
		mu:            &sync.Mutex{},
		NotifyExpired: make(chan K),
		records:       make(map[K]*Record[T]),
	}

	go cache.ListenExpired()

	return cache
}

// get value if it exists
// if not then return ok false
func (c *Cache[K, T]) Take(key K) (T, bool) {
	c.mu.Lock()

	record, ok := c.records[key]
	if !ok {
		c.mu.Unlock()
		return *new(T), false
	}

	ok = record.in_use.TryLock()
	if ok {
		c.mu.Unlock()
		return record.value, true
	}

	record.waiting.Add(1)
	c.mu.Unlock()

	// wait until
	record.in_use.Lock()
	record.waiting.Add(-1)

	return record.value, true
}

func (c *Cache[K, T]) Return(key K) {
	c.mu.Lock()

	record, ok := c.records[key]
	if !ok {
		c.mu.Unlock()
		return
	}

	record.in_use.TryLock()
	record.in_use.Unlock()

	c.mu.Unlock()
}

func (c *Cache[K, T]) CreateIfNotExistedAndTake(key K, value T) (v T, was_created bool) {
	c.mu.Lock()

	record, ok := c.records[key]
	if !ok {
		record := &Record[T]{
			in_use:  &sync.Mutex{},
			waiting: &atomic.Int32{},

			value: value,
		}

		c.records[key] = record
		record.in_use.Lock()
		c.mu.Unlock()

		return record.value, true
	}

	ok = record.in_use.TryLock()
	if ok {
		c.mu.Unlock()
		return record.value, false
	}

	record.waiting.Add(1)
	c.mu.Unlock()

	record.in_use.Lock()
	record.waiting.Add(-1)

	return record.value, false
}

func (c *Cache[K, T]) ListenExpired() {
	for {
		key := <-c.NotifyExpired

		c.mu.Lock()
		record, ok := c.records[key]
		if !ok {
			c.mu.Unlock()
			continue
		}

		ok = record.in_use.TryLock()
		if ok && record.waiting.Load() == 0 {
			delete(c.records, key)
			c.mu.Unlock()
			continue
		}
		// if it is busy, there is no point to delete it
		c.mu.Unlock()
	}
}
