package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	. "github.com/aandryashin/matchers"
	. "github.com/aandryashin/matchers/httpresp"
)

type MapStore struct {
	l sync.RWMutex
	i uint64
	m map[uint64]string
}

func NewMapStore() *MapStore {
	return &MapStore{m: make(map[uint64]string)}
}

func (store *MapStore) Get(k uint64) (string, bool) {
	store.l.RLock()
	defer store.l.RUnlock()
	v, ok := store.m[k]
	return v, ok
}

func (store *MapStore) Put(v string) uint64 {
	store.l.Lock()
	defer store.l.Unlock()
	store.i++
	store.m[store.i] = v
	return store.i
}

type BrokenStore struct{}

func (store *BrokenStore) Get(k uint64) (string, bool) {
	panic(errors.New("store is broken"))
}

func (store *BrokenStore) Put(v string) uint64 {
	panic(errors.New("store is broken"))
}

var (
	srv *httptest.Server
)

func init() {
	store = NewMapStore()
	srv = httptest.NewServer(handler())
}

func uri(s string) string {
	return srv.URL + s
}

func body(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
}

func TestIndex(t *testing.T) {
	resp, err := http.Get(uri("/"))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusOK})
}

func TestWrongMethod(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPut, uri("/"), nil)
	resp, err := http.DefaultClient.Do(req)
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusBadRequest})
}

func TestBadRequest(t *testing.T) {
	params := []string{
		`{`,
		`{}`,
		`{"url" : ""}`,
		`{"url" : "https://"}`,
		`{"url" : "example.com"}`,
	}
	for _, param := range params {
		resp, err := http.Post(uri("/"), "", body(param))
		AssertThat(t, err, Is{nil})
		AssertThat(t, resp, Code{http.StatusBadRequest})
	}
}

func TestBadKey(t *testing.T) {
	resp, err := http.Get(uri("/ "))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusBadRequest})
}

func TestMissingKey(t *testing.T) {
	resp, err := http.Get(uri("/missing"))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusNotFound})
}

func TestRedirect(t *testing.T) {
	k := store.Put("url")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	resp, err := client.Get(uri("/" + encode(k)))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusMovedPermanently})

	loc, err := resp.Location()
	AssertThat(t, err, Is{nil})
	AssertThat(t, loc.String(), EqualTo{uri("/url")})
}

func TestBrokenStoreGet(t *testing.T) {
	store = &BrokenStore{}
	resp, err := http.Get(uri("/key"))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusInternalServerError})
}

func TestBrokenStorePost(t *testing.T) {
	store = &BrokenStore{}
	resp, err := http.Post(uri("/"), "", body(`{"url" : "http://example.com"}`))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusInternalServerError})
}
