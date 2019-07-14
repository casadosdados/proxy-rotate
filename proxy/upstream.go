package proxy

//reference: https://github.com/elazarl/goproxy/issues/201

import (
	"context"
	"github.com/elazarl/goproxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	"github.com/casadosdados/proxy-rotate/util"
	"github.com/casadosdados/socks"
)


var proxys []string
var ProxyList *ProxyBucket
var proxyCacheIgnore = NewProxyCacheIgnore()




func forward(proxy *goproxy.ProxyHttpServer) func(network, addr string) (net.Conn, error)  {

	dial := func(network, addr string) (net.Conn, error) {
		// Prevent upstream proxy from being re-directed
		if i, k := proxyCacheIgnore.Load(addr); i && k {
			return net.Dial(network, addr)
		}

		toAddr := ProxyList.RandomProxy().Parse()
		u, err := url.Parse(toAddr)
		if err != nil {
			log.Fatal("failed to parse upstream server:", err)
		}

		if u.Scheme == "socks5" || u.Scheme == "socks4"{
			dialSocks := socks.Dial(toAddr)
			return dialSocks(network, addr)
		}

		dialer := proxy.NewConnectDialToProxy(toAddr)
		if dialer == nil {
			panic("nil dialer, invalid uri?")
		}
		return dialer(network, addr)
	}
	return dial
}

func forwardContext(proxy *goproxy.ProxyHttpServer) func(ctx2 context.Context, network, addr string) (net.Conn, error)  {

	dial := func(ctx2 context.Context, network, addr string) (net.Conn, error) {
		// Prevent upstream proxy from being re-directed
		if i, k := proxyCacheIgnore.Load(addr); i && k {
			return net.Dial(network, addr)
		}

		toAddr := ProxyList.RandomProxy().Parse()
		u, err := url.Parse(toAddr)
		if err != nil {
			log.Fatal("failed to parse upstream server:", err)
		}

		if u.Scheme == "socks5" || u.Scheme == "socks4"{
			dialSocks := socks.Dial(toAddr)
			return dialSocks(network, addr)
		}

		dialer := proxy.NewConnectDialToProxy(toAddr)
		if dialer == nil {
			panic("nil dialer, invalid uri?")
		}
		return dialer(network, addr)
	}
	return dial
}

func NewTransport (proxy *goproxy.ProxyHttpServer) *http.Transport {
	return &http.Transport{
		DialContext:    		forwardContext(proxy),
		Dial:    				forward(proxy),
		DisableKeepAlives:		true,
		IdleConnTimeout:		90 * time.Second,
		TLSHandshakeTimeout:	10 * time.Second,
		ExpectContinueTimeout: 	1 * time.Second,
		MaxIdleConns:			100,
	}
}



func (pb *ProxyBucket) RandomProxy() *Proxy {
	lenProxys := len(pb.Proxy)
	if lenProxys == 0 {
		time.Sleep(5 * time.Second)
		return pb.RandomProxy()
	}
	index := util.RandInt(0, lenProxys)
	p := pb.Proxy[index]

	//log.Println("proxy returned", index, p)
	return p
}


type ProxyCacheIgnore struct {
	sync.RWMutex
	internal map[string]bool
}

func NewProxyCacheIgnore() *ProxyCacheIgnore {
	return &ProxyCacheIgnore{
		internal: make(map[string]bool),
	}
}

func (rm *ProxyCacheIgnore) Load(key string) (bool, bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *ProxyCacheIgnore) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *ProxyCacheIgnore) Store(key string, value bool) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}