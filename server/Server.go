package server

import (
	"fmt"
	. "github.com/Thebigwind/go-project/common"
	"net"
	"net/http"
	"sync"
)

type XtRestServer struct {
	addr string
}

type limitListener struct {
	net.Listener
	sem chan struct{}  // 利用chan的缓存队列机制来限制连接
}

var GlobalRestServer *XtRestServer = nil

func NewRESTServer(addr string) *XtRestServer {
	if addr == "" {
		addr = ":http"
	}
	GlobalRestServer = &XtRestServer{
		addr: addr,
	}
	return GlobalRestServer
}

func (server *XtRestServer) StartRESTServer(max int) {
	if max <= 0 {
		max = 1000
	}
	router := NewRouter()

	ln, err := net.Listen("tcp", server.addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	listener := LimitListener(ln, uint(max))
	Logger.Fatal(http.Serve(listener, router))
}

//////////////////////////////////////////////////////////

func LimitListener(l net.Listener, max uint) net.Listener {
	return &limitListener{l, make(chan struct{}, max)}
}

func (l *limitListener) acquire() { l.sem <- struct{}{} }
func (l *limitListener) release() { <-l.sem }

func (l *limitListener) Accept() (net.Conn, error) {
	l.acquire()
	conn, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	return &limitListenerConn{Conn: conn, release: l.release}, nil
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}