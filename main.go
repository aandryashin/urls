package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/facebookgo/grace/gracehttp"
)

var (
	index     string
	html      []byte
	folder    string
	endpoints []string
	httpAddr  string
	httpsAddr string
	sslCert   string
	sslKey    string
	store     Store
)

func redirect(port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostport := r.Context().Value(http.LocalAddrContextKey).(net.Addr).String()
		host, _, _ := net.SplitHostPort(hostport)
		r.URL.Scheme = "https"
		r.URL.Host = net.JoinHostPort(host, port)
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	})
}

func init() {
	var list string
	flag.StringVar(&index, "index", "html/index.html", "index to serve on root")
	flag.StringVar(&folder, "folder", "/", "etcd folder to store keys")
	flag.StringVar(&list, "endpoints", "http://127.0.0.1:2379", "comma-separated list of etcd endpoints")
	flag.StringVar(&httpAddr, "http", ":8080", "listen http protocol on")
	flag.StringVar(&httpsAddr, "https", "", "listen https protocol on (default disabled)")
	flag.StringVar(&sslCert, "ssl-cert", "cert/localhost.crt", "ssl certificate")
	flag.StringVar(&sslKey, "ssl-key", "cert/localhost.key", "ssl certificate key")
	flag.Parse()
	endpoints = strings.Split(list, ",")
}

func main() {
	var err error
	html, err = ioutil.ReadFile(index)
	if err != nil {
		log.Fatal(err)
	}
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	store = NewEtcdStore(folder, c)

	servers := []*http.Server{}
	if httpsAddr == "" {
		servers = append(servers, &http.Server{
			Addr:    httpAddr,
			Handler: handler(),
		})
	} else {
		_, httpsPort, err := net.SplitHostPort(httpsAddr)
		if err != nil {
			log.Fatal("malformed address", httpAddr)
		}
		cer, err := tls.LoadX509KeyPair(sslCert, sslKey)
		if err != nil {
			log.Fatal("error load certificates")
		}
		servers = append(servers, &http.Server{
			Addr:    httpAddr,
			Handler: redirect(httpsPort),
		})
		servers = append(servers, &http.Server{
			Addr:      httpsAddr,
			Handler:   handler(),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cer}},
		})
	}
	log.Fatal(gracehttp.Serve(servers...))
}
