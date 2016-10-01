package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type jso struct {
	Url string `json:"url"`
}

var (
	store Store
)

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
			k := store.Put(o.Url)
			json.NewEncoder(w).Encode(jso{k})
		case http.MethodGet:
			k := strings.TrimPrefix(r.URL.RequestURI(), "/")
			v, ok := store.Get(k)
			if !ok {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			http.Redirect(w, r, v, http.StatusMovedPermanently)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	})
	return mux
}
