package main

import (
	"strconv"
	"sync"
)

type Store interface {
	Put(k string) string
	Get(v string) (string, bool)
}

type MapStore struct {
	l sync.RWMutex
	i uint64
	m map[uint64]string
}

func NewMapStore() *MapStore {
	return &MapStore{m: make(map[uint64]string)}
}

func (store *MapStore) Get(k string) (string, bool) {
	store.l.RLock()
	defer store.l.RUnlock()
	i, err := decode(k)
	if err != nil {
		return "", false
	}
	v, ok := store.m[i]
	return v, ok
}

func (store *MapStore) Put(v string) string {
	store.l.Lock()
	defer store.l.Unlock()
	store.i++
	store.m[store.i] = v
	return encode(store.i)
}

func decode(s string) (uint64, error) {
	return strconv.ParseUint(s, 36, 64)
}

func encode(i uint64) string {
	return strconv.FormatUint(i, 36)
}
