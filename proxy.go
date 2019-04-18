package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Proxy struct {
	PublicPort int
	TunnelPort int
	Tunnel     net.Conn
}

func NewProxy(publicPort, tunnelPort int) *Proxy {
	return &Proxy{PublicPort: publicPort, TunnelPort: tunnelPort}
}

func listen(addr string) net.Listener {
	l, err := net.Listen("tcp", addr)
	check(err, "proxy.listen.start")
	return l
}

func (p *Proxy) forward(from net.Conn, to net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, _ = io.Copy(from, to)
		wg.Done()
	}()
	go func() {
		_, _ = io.Copy(from, to)
		wg.Done()
	}()
	wg.Wait()
}

func (p *Proxy) Start() {
	go func() {
		l := listen(fmt.Sprintf(":%d", p.PublicPort))
		for {
			conn, err := l.Accept()
			check(err, "proxy.listen.accept")
			go p.forward(conn, p.Tunnel)
		}
	}()
	go func() {
		l := listen(fmt.Sprintf(":%d", p.Tunnel))
		for {
			conn, err := l.Accept()
			if !check(err, "proxy.listen.accept") {
				if p.Tunnel != nil {
					_ = p.Tunnel.Close()
				}
				p.Tunnel = conn
			}
			if conn, ok := conn.(*net.TCPConn); ok {
				check(conn.SetKeepAlive(true))
			}
		}
	}()
}
