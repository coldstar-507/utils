package utils

import "sync"

type Smap[K comparable, V any] struct {
	m sync.RWMutex
	M map[K]V
}

func (sm *Smap[K, V]) RLock() {
	sm.m.RLock()
}

func (sm *Smap[K, V]) RUnlock() {
	sm.m.RUnlock()
}

func (sm *Smap[K, V]) Lock() {
	sm.m.Lock()
}

func (sm *Smap[K, V]) Unlock() {
	sm.m.Unlock()
}

func (sm *Smap[K, V]) Read(key K) (V, bool) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	v, ok := sm.M[key]
	return v, ok
}

func (sm *Smap[K, V]) Swap(key K, value V) (V, bool) {
	sm.m.Lock()
	defer sm.m.Unlock()
	cur, ok := sm.M[key]
	sm.M[key] = value
	if ok {
		return cur, true
	}
	return cur, false
}

func (sm *Smap[K, V]) SwapIf(key K, value V, cond func(value V, cur V) bool) (V, bool) {
	sm.m.Lock()
	defer sm.m.Unlock()
	cur, ok := sm.M[key]
	if !ok {
		sm.M[key] = value
		return cur, false
	} else if cond(value, cur) {
		sm.M[key] = value
		return cur, true
	} else {
		return cur, false
	}
}

func (sm *Smap[K, V]) Delete(key K) int {
	sm.m.Lock()
	defer sm.m.Unlock()
	delete(sm.M, key)
	return len(sm.M)
}

func (sm *Smap[K, V]) DeleteIf(key K, f func(value V) bool) (int, bool) {
	sm.m.Lock()
	defer sm.m.Unlock()
	v, ok := sm.M[key]
	if !ok || !f(v) {
		return len(sm.M), false
	} else {
		delete(sm.M, key)
		return len(sm.M), true
	}
}

func (sm *Smap[K, V]) DeleteAll(keys []K) {
	sm.m.Lock()
	defer sm.m.Unlock()
	for _, key := range keys {
		delete(sm.M, key)
	}
}

func (sm *Smap[K, V]) Do(f func(key K, value V)) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	for k, v := range sm.M {
		f(k, v)
	}
}

func (sm *Smap[K, V]) Dok(f func(key K)) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	for k := range sm.M {
		f(k)
	}
}

func (sm *Smap[K, V]) Dov(f func(value V)) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	for _, v := range sm.M {
		f(v)
	}
}

func (sm *Smap[K, V]) GoDov(s chan struct{}, wg *sync.WaitGroup, f func(value V)) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	for _, v := range sm.M {
		s <- struct{}{}
		wg.Add(1)
		go f(v)
	}
}

func (sm *Smap[K, V]) Modifying(f func(e map[K]V)) {
	sm.m.Lock()
	defer sm.m.Unlock()
	f(sm.M)
}

func (sm *Smap[K, V]) ModifyingAt(key K, isThere func(V), isNotThere func() V) {
	sm.m.Lock()
	defer sm.m.Unlock()
	v, ok := sm.M[key]
	if ok {
		isThere(v)
	} else if isNotThere != nil {
		sm.M[key] = isNotThere()
	}
}

func (sm *Smap[K, V]) Reading(f func(e map[K]V)) {
	sm.m.RLock()
	defer sm.m.RUnlock()
	f(sm.M)
}

func (sm *Smap[K, V]) ReadingAt(key K, f func(e V)) bool {
	sm.m.RLock()
	defer sm.m.RUnlock()
	e, ok := sm.M[key]
	if !ok {
		return false
	}
	f(e)
	return true
}
