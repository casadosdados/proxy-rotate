### Proxy Rotate
Proxy server using [goproxy](https://github.com/elazarl/goproxy) and [Proxy List](https://www.proxy-list.download/) for rotate proxy

#### Support filter by country
add env: COUNTRY=br (case insensitive)

#### Run with docker
```bash
docker run --name proxy-rotate -p 8888:8888 -e COUNTRY=br -d casadosdados/proxy-rotate
```

#### With Golang
```bash
go run main/proxy
```
Default port: 8888

Connecting: https://localhost:8888