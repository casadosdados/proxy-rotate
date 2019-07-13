package proxy

//reference: https://github.com/elazarl/goproxy/issues/201

import (
	"github.com/elazarl/goproxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	"github.com/casadosdados/proxy-rotate/util"
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
		_, err := url.Parse(toAddr)
		if err != nil {
			log.Fatal("failed to parse upstream server:", err)
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
		//DialContext:    forward(proxy),
		Dial:    forward(proxy),
		DisableKeepAlives:true,
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