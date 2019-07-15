package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"net/http"
	proxy2 "github.com/casadosdados/proxy-rotate/proxy"
	"log"
	"time"
)

func main() {
	var host string
	flag.StringVar(&host, "h", "0.0.0.0:8888", "host and port")
	flag.Parse()

	httpServer := &http.Server{
		Addr:         host,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	httpServer.SetKeepAlivesEnabled(false)
	proxy2.ProxyList = new(proxy2.ProxyBucket)
	go proxy2.ProxyList.Start()
	fmt.Println("Starting proxy server ", host)
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	proxy.Tr = proxy2.NewTransport(proxy)

	httpServer.Handler = proxy
	log.Fatal(httpServer.ListenAndServe())
}