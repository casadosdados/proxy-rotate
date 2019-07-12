package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"net/http"
	proxy2 "../proxy"
	"log"
)

func main() {
	proxy2.ProxyList = new(proxy2.ProxyBucket)
	go proxy2.ProxyList.Start()
	fmt.Println("Starting proxy server...")
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	proxy.Tr = proxy2.NewTransport(proxy)

	log.Fatal(http.ListenAndServe("0.0.0.0:8888", proxy))
}