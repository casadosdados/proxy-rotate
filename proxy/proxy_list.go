package proxy

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

)

var URL_PROXY_LIST = map[string]string{
	"http": "https://www.proxy-list.download/api/v1/get?type=http",
	"socks4": "https://www.proxy-list.download/api/v1/get?type=socks4",
	"socks5": "https://www.proxy-list.download/api/v1/get?type=socks5",
}
const URL_API_IP = "http://ip-api.com/json/"
const WORKERS_PROXY_LIST = 10



type Proxy struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Schema string `json:"schema"`
	Info struct {
		AS string `json:"as"`
		City string `json:"city"`
		Country string `json:"country"`
		ContryCode string `json:"contry_code"`
		Isp string `json:"isp"`
		Query string `json:"query"`
		Region string `json:"region"`
		RegionName string `json:"region_name"`
		Status string `json:"status"`
		Timezone string `json:"timezone"`
		Zip string `json:"zip"`
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
}

type ProxyBucket struct {
	Proxy []*Proxy
}

func (p *Proxy) Check() error {
	resp, _, errs := gorequest.New().Get(URL_API_IP).
		Proxy(p.Parse()).
		Timeout(10 * time.Second).
		EndStruct(&p.Info)
	if len(errs) > 0 {
		return errs[0]
	}
	if resp.StatusCode != 200 {
		return errors.New("status code " + resp.Status)
	}
	if p.Info.Status != "success" {
		return errors.New("status response " + p.Info.Status)
	}
	return nil
}

func (p *Proxy) Parse() string {
	return fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
}

func (p *ProxyBucket) Start() {
	for {
		for typeProxy, urlProxy := range URL_PROXY_LIST {

			_, resp, errs := gorequest.New().Timeout(20 * time.Second).Get(urlProxy).End()
			if len(errs) > 0 {
				log.Println(errs)
				time.Sleep(10 * time.Second)
				return
			}
			ch := make(chan *Proxy, 10000)
			proxys := strings.Split(resp, "\n")

			log.Println(len(proxys), "proxy found", urlProxy)

			for _, proxy := range proxys {
				proxy = strings.Replace(proxy, "\r", "", 2)
				hostPort := strings.Split(proxy, ":")
				if len(hostPort) != 2 {
					continue
				}
				port, err := strconv.Atoi(hostPort[1])
				if err != nil {
					log.Println("error on atoi port", err)
					continue
				}
				newProxy := &Proxy{
					Schema: typeProxy,
					Host: hostPort[0],
					Port: port,
				}
				ch <- newProxy
			}
			close(ch)
			for i:=0; i<=WORKERS_PROXY_LIST; i++ {
				go p.newProxy(ch)
			}
		}
		time.Sleep(2 * time.Hour)
	}
}

func (p *ProxyBucket) newProxy(ch chan *Proxy) {
	for proxy := range ch {
		if proxy.Schema != "socks4" {
			if err:= proxy.Check(); err != nil {
				//log.Println("error on check", proxy.Parse(), err)
				continue
			}
		}
		//proxy.Schema = "socks5"
		//log.Println("append new proxy", proxy)
		proxyCacheIgnore.Store(fmt.Sprintf("%s:%d", proxy.Host, proxy.Port), true)
		if len(p.Proxy) >= 20000 {
			p.Proxy = p.Proxy[1:]
		}
		p.Proxy = append(p.Proxy, proxy)
	}
}

