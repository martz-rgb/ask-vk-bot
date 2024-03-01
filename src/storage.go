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

type Storage[K comparable, T interface{}] struct {
	mu *sync.Mutex

	NotifyExpired chan K
	records       map[K]*Record[T]
}

func NewStorage[K comparable, T interface{}]() *Storage[K, T] {
	storage := &Storage[K, T]{
		mu:            &sync.Mutex{},
		NotifyExpired: make(chan K),
		records:       make(map[K]*Record[T]),
	}

	go storage.ListenExpired()

	return storage
}

// get value if it exists
// if not then return ok false
func (s *Storage[K, T]) Take(key K) (T, bool) {
	s.mu.Lock()

	record, ok := s.records[key]
	if !ok {
		s.mu.Unlock()
		return *new(T), false
	}

	ok = record.in_use.TryLock()
	if ok {
		s.mu.Unlock()
		return record.value, true
	}

	record.waiting.Add(1)
	s.mu.Unlock()

	// wait until
	record.in_use.Lock()
	record.waiting.Add(-1)

	return record.value, true
}

func (s *Storage[K, T]) Return(key K) {
	s.mu.Lock()

	record, ok := s.records[key]
	if !ok {
		s.mu.Unlock()
		return
	}

	record.in_use.TryLock()
	record.in_use.Unlock()

	s.mu.Unlock()
}

func (s *Storage[K, T]) CreateIfNotExistedAndTake(key K, value T) (v T, existed bool) {
	s.mu.Lock()

	record, ok := s.records[key]
	if !ok {
		record := &Record[T]{
			in_use:  &sync.Mutex{},
			waiting: &atomic.Int32{},

			value: value,
		}

		s.records[key] = record
		record.in_use.Lock()
		s.mu.Unlock()

		return record.value, false
	}

	ok = record.in_use.TryLock()
	if ok {
		s.mu.Unlock()
		return record.value, true
	}

	record.waiting.Add(1)
	s.mu.Unlock()

	record.in_use.Lock()
	record.waiting.Add(-1)

	return record.value, true
}

func (s *Storage[K, T]) ListenExpired() {
	for {
		key := <-s.NotifyExpired

		s.mu.Lock()
		record, ok := s.records[key]
		if !ok {
			s.mu.Unlock()
			continue
		}

		ok = record.in_use.TryLock()
		if ok && record.waiting.Load() == 0 {
			delete(s.records, key)
			s.mu.Unlock()
			continue
		}
		// if it is busy, there is no point to delete it
		s.mu.Unlock()
	}
}
