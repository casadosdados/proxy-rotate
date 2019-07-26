FROM golang:1.12 as builder
WORKDIR /proxy
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o proxy-rotate main/proxy.go

FROM scratch
WORKDIR /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /proxy/proxy-rotate ./
ENTRYPOINT ["./proxy-rotate"]
