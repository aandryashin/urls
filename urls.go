package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type jso struct {
	Url string `json:"url"`
}

var (
	mlock  sync.RWMutex
	urls   map[uint64]string = make(map[uint64]string)
	lock   sync.Mutex
	serial uint64
)

func increment() uint64 {
	defer func() {
		lock.Lock()
		serial++
		lock.Unlock()
	}()
	return serial
}

func decode(s string) (uint64, error) {
	return strconv.ParseUint(s, 36, 64)
}

func encode(i uint64) string {
	return strconv.FormatUint(i, 36)
}

func valid(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.IsAbs() && u.Host != ""
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/" && r.Method == http.MethodGet {
			// TODO: write html
			return
		}
		switch r.Method {
		case http.MethodPost:
			var o jso
			json.NewDecoder(r.Body).Decode(&o)
			r.Body.Close()
			if !valid(o.Url) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			key := increment()
			mlock.Lock()
			urls[key] = o.Url
			mlock.Unlock()
			json.NewEncoder(w).Encode(jso{encode(key)})
		case http.MethodGet:
			uri := strings.TrimPrefix(r.URL.RequestURI(), "/")
			key, err := decode(uri)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			mlock.RLock()
			ru, ok := urls[key]
			mlock.RUnlock()
			if !ok {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			http.Redirect(w, r, ru, http.StatusMovedPermanently)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	})
	return mux
}
