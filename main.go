package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
)

var (
	httpAddr  string
	httpsAddr string
	sslCert   string
	sslKey    string
)

func redirectTo(port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostport := r.Context().Value(http.LocalAddrContextKey).(net.Addr).String()
		host, _, _ := net.SplitHostPort(hostport)
		u := fmt.Sprintf("https://%s:%s", host, port)
		http.Redirect(w, r, u, http.StatusMovedPermanently)
	})
}

func init() {
	flag.StringVar(&httpAddr, "http", ":8080", "listen http protocol on")
	flag.StringVar(&httpsAddr, "https", "", "listen https protocol on (default disabled)")
	flag.StringVar(&sslCert, "ssl-cert", "cert/localhost.crt", "ssl certificate")
	flag.StringVar(&sslKey, "ssl-key", "cert/localhost.key", "ssl certificate key")
	flag.Parse()
}

func main() {
	store = NewMapStore()
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
			Handler: redirectTo(httpsPort),
		})
		servers = append(servers, &http.Server{
			Addr:      httpsAddr,
			Handler:   handler(),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cer}},
		})
	}
	log.Fatal(gracehttp.Serve(servers...))
}
