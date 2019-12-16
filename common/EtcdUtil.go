package common

import (
	"github.com/coreos/etcd/client"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	ETCD_TRANSPORT_KEEPALIVE time.Duration = 10 * time.Second
	ETCD_CONNECT_TIMEOUT     time.Duration = 10 * time.Second
	ETCD_TLS_HANDSHK_TIMEOUT time.Duration = 10 * time.Second
	ETCD_HEADER_TIMEOUT      time.Duration = 5 * time.Second
	ETCD_CTX_TIMEOUT         time.Duration = 20 * time.Second
)

var Transport client.CancelableTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   ETCD_CONNECT_TIMEOUT,
		KeepAlive: ETCD_TRANSPORT_KEEPALIVE,
	}).Dial,
	TLSHandshakeTimeout: ETCD_TLS_HANDSHK_TIMEOUT,
}

type PowerContext struct {
	Endpoints []string
	Kapi      client.KeysAPI
}

var GlobalContext *PowerContext = nil

func InitGlobalContext(endpoints []string) error {
	GlobalContext = &PowerContext{}

	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               Transport,
		HeaderTimeoutPerRequest: ETCD_HEADER_TIMEOUT,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
		return err
	}

	GlobalContext.Endpoints = endpoints
	kAPI := client.NewKeysAPI(c)
	GlobalContext.Kapi = kAPI

	return nil
}

func GetGlobalContext() *PowerContext {
	return GlobalContext
}
