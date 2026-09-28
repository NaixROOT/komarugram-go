package set

import (
	"sync"
	"sync/atomic"
)

type SyncSet[T comparable] struct {
	list    sync.Map //empty structs occupy 0 memory
	counter atomic.Int64
}

func (s *SyncSet[T]) Range(f func(key any, value any) bool) {
	s.list.Range(f)
}

func (s *SyncSet[T]) Has(v T) bool {
	_, ok := s.list.Load(v)
	return ok
}

func (s *SyncSet[T]) Add(v T) {
	s.list.Store(v, struct{}{})
	s.counter.Add(1)
}

func (s *SyncSet[T]) Delete(v T) {
	_, ok := s.list.LoadAndDelete(v)
	if ok {
		s.counter.Add(-1)
	}
}

func (s *SyncSet[T]) Remove(v T) {
	s.Delete(v)
}

func (s *SyncSet[T]) Clear() {
	s.list.Clear()
	s.counter.Store(0)
}

func (s *SyncSet[T]) Size() int64 {
	return s.counter.Load()
}

func (s *SyncSet[T]) List() []T {
	list := make([]T, s.Size())
	s.Range(func(key, value any) bool {
		list = append(list, key.(T))
		return true
	})
	return list
}

func NewSyncSet[T comparable]() *SyncSet[T] {
	s := &SyncSet[T]{}
	return s
}

// AddMulti Add multiple values in the set
func (s *SyncSet[T]) AddMulti(list ...T) {
	for _, v := range list {
		s.list.Store(v, struct{}{})
	}
}

// Filter returns a subset, that contains only the values that satisfies the given predicate P
func (s *SyncSet[T]) Filter(P FilterFunc[T]) *Set[T] {
	res := &Set[T]{}
	res.list = make(map[T]struct{})
	for v := range s.list.Range {
		if !P(v.(T)) {
			continue
		}
		res.Add(v.(T))
	}
	return res
}

func (s *SyncSet[T]) Union(s2 *SyncSet[T]) *SyncSet[T] {
	res := NewSyncSet[T]()
	for v := range s.list.Range {
		res.Add(v.(T))
	}

	for v := range s2.list.Range {
		res.Add(v.(T))
	}
	return res
}

func (s *SyncSet[T]) Intersect(s2 *SyncSet[T]) *SyncSet[T] {
	res := NewSyncSet[T]()
	for v := range s.list.Range {
		if !s2.Has(v.(T)) {
			continue
		}
		res.Add(v.(T))
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *SyncSet[T]) Difference(s2 *SyncSet[T]) *SyncSet[T] {
	res := NewSyncSet[T]()
	for v := range s.list.Range {
		if s2.Has(v.(T)) {
			continue
		}
		res.Add(v.(T))
	}
	return res
}
