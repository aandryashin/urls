package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type jso struct {
	Url string `json:"url"`
}

func valid(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.IsAbs() && u.Host != ""
}

func decode(s string) (uint64, error) {
	return strconv.ParseUint(s, 36, 64)
}

func encode(i uint64) string {
	return strconv.FormatUint(i, 36)
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			e := recover()
			if e != nil {
				log.Println(e)
				http.Error(w, e.(error).Error(), http.StatusInternalServerError)
			}
		}()
		if r.URL.RequestURI() == "/" && r.Method == http.MethodGet {
			buf, err := ioutil.ReadFile(index)
			if err != nil {
				panic(err)
			}
			w.Write(buf)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var o jso
			json.NewDecoder(r.Body).Decode(&o)
			r.Body.Close()
			if !valid(o.Url) {
				http.Error(w, fmt.Sprintf("Bad url: %s", o.Url), http.StatusBadRequest)
				return
			}
			k := store.Put(o.Url)
			json.NewEncoder(w).Encode(jso{encode(k)})
		case http.MethodGet:
			p := strings.TrimPrefix(r.URL.RequestURI(), "/")
			k, err := decode(p)
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad key encoding: %s", err), http.StatusBadRequest)
				return
			}
			v, ok := store.Get(k)
			if !ok {
				http.Error(w, fmt.Sprintf("Url not found: %s", p), http.StatusNotFound)
				return
			}
			http.Redirect(w, r, v, http.StatusMovedPermanently)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	})
	return mux
}
