package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/aandryashin/matchers"
	. "github.com/aandryashin/matchers/httpresp"
)

var (
	srv *httptest.Server
)

func init() {
	srv = httptest.NewServer(handler())
}

func uri(s string) string {
	return srv.URL + s
}

func body(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
}

func TestRootGet(t *testing.T) {
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
	serial := increment()
	mlock.Lock()
	urls[serial] = "url"
	mlock.Unlock()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	resp, err := client.Get(uri("/" + encode(serial)))
	AssertThat(t, err, Is{nil})
	AssertThat(t, resp, Code{http.StatusMovedPermanently})

	loc, err := resp.Location()
	AssertThat(t, err, Is{nil})
	AssertThat(t, loc.String(), EqualTo{uri("/url")})
}

func TestNewKey(t *testing.T) {
	resp, err := http.Post(uri("/"), "", body(`{"url" : "http://example.com"}`))
	AssertThat(t, err, Is{nil})
	var o jso
	AssertThat(t, resp, AllOf{Code{http.StatusOK}, IsJson{&o}})

	k, err := decode(strings.TrimPrefix(o.Url, "/"))
	AssertThat(t, err, Is{nil})

	mlock.RLock()
	u, ok := urls[k]
	mlock.RUnlock()

	AssertThat(t, ok, Is{true})
	AssertThat(t, u, EqualTo{"http://example.com"})
}
