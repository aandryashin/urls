package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/client"
)

type Store interface {
	Put(v string) uint64
	Get(k uint64) (string, bool)
}

type EtcdStore struct {
	dir string
	api client.KeysAPI
}

func NewEtcdStore(d string, c client.Client) *EtcdStore {
	return &EtcdStore{d, client.NewKeysAPI(c)}
}

func (store *EtcdStore) Get(k uint64) (string, bool) {
	p := fmt.Sprintf("%s/%020d", store.dir, k)
	resp, err := store.api.Get(context.Background(), p, nil)
	if err != nil {
		if client.IsKeyNotFound(err) {
			return "", false
		}
		panic(err)
	}
	return resp.Node.Value, true
}

func (store *EtcdStore) Put(v string) uint64 {
	resp, err := store.api.CreateInOrder(context.Background(), store.dir, v, nil)
	if err != nil {
		panic(err)
	}
	return resp.Index
}
