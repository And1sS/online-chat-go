package util

import (
	"math/rand"
	"sync"
)

type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

func (s *SafeMap[K, V]) Set(k K, v V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[k] = v
}

func (s *SafeMap[K, V]) ComputeIfAbsent(k K, supplier func() V) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.data[k]

	if !ok {
		val = supplier()
		s.data[k] = val
	}

	return val, !ok
}

func (s *SafeMap[K, V]) Get(k K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[k]
	return val, ok
}

func (s *SafeMap[K, V]) ForRandomEntry(f func(K, V)) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx := 0
	size := len(s.data)
	if size == 0 {
		return false
	}

	target := rand.Int() % size

	for key, val := range s.data {
		if idx == target {
			f(key, val)
			return true
		}
		idx++
	}
	return false
}

func (s *SafeMap[K, V]) Delete(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, k)
}

func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *SafeMap[K, V]) ForEach(f func(K, V)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for key, val := range s.data {
		f(key, val)
	}
}
