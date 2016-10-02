package main

import (
	"errors"
	"testing"

	. "github.com/aandryashin/matchers"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type apiGetError struct {
	client.KeysAPI
}

func (api apiGetError) Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error) {
	return nil, errors.New("store is broken")
}

func TestGetPanic(t *testing.T) {
	api := apiGetError{}
	store := &EtcdStore{"/", client.KeysAPI(api)}
	defer func() {
		e := recover()
		AssertThat(t, e.(error).Error(), EqualTo{"store is broken"})
	}()
	store.Get(0)
}

type apiGetErrorKeyNotFound struct {
	client.KeysAPI
}

func (api apiGetErrorKeyNotFound) Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error) {
	return nil, client.Error{Code: client.ErrorCodeKeyNotFound}
}

func TestGetKeyNotFound(t *testing.T) {
	api := apiGetErrorKeyNotFound{}
	store := &EtcdStore{"/", client.KeysAPI(api)}
	_, ok := store.Get(0)
	AssertThat(t, ok, Is{false})
}

type apiGetValue struct {
	client.KeysAPI
}

func (api apiGetValue) Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error) {
	return &client.Response{Node: &client.Node{Value: "value"}}, nil
}

func TestGet(t *testing.T) {
	api := apiGetValue{}
	store := &EtcdStore{"/", client.KeysAPI(api)}
	v, ok := store.Get(0)
	AssertThat(t, ok, Is{true})
	AssertThat(t, v, EqualTo{"value"})
}

type apiCreateInOrderError struct {
	client.KeysAPI
}

func (api apiCreateInOrderError) CreateInOrder(ctx context.Context, dir, value string, opts *client.CreateInOrderOptions) (*client.Response, error) {
	return nil, errors.New("store is broken")
}

func TestPutError(t *testing.T) {
	api := apiCreateInOrderError{}
	store := &EtcdStore{"/", client.KeysAPI(api)}
	defer func() {
		e := recover()
		AssertThat(t, e.(error).Error(), EqualTo{"store is broken"})
	}()
	store.Put("")
}

type apiCreateInOrderValue struct {
	client.KeysAPI
}

func (api apiCreateInOrderValue) CreateInOrder(ctx context.Context, dir, value string, opts *client.CreateInOrderOptions) (*client.Response, error) {
	return &client.Response{Index: 12345}, nil
}

func TestPutValue(t *testing.T) {
	api := apiCreateInOrderValue{}
	store := &EtcdStore{"/", client.KeysAPI(api)}
	k := store.Put("")
	AssertThat(t, k, EqualTo{uint64(12345)})
}
