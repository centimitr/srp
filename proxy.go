package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Proxy struct {
	PublicAddr string
	TunnelAddr string
	Tunnel     net.Conn
}

func NewProxy(publicAddr, tunnelAddr string) *Proxy {
	return &Proxy{PublicAddr: publicAddr, TunnelAddr: tunnelAddr}
}

func copyTCPBuf(from net.Conn, to net.Conn, wg *sync.WaitGroup) {
	_, err := io.Copy(from, to)
	check(err)
	//err = from.Close()
	//check(err)
	wg.Done()
}

func (p *Proxy) forward(from net.Conn, to net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go copyTCPBuf(to, from, &wg)
	go copyTCPBuf(from, to, &wg)
	wg.Wait()
	fmt.Println("DONE")
}

func (p *Proxy) Start() (err error) {
	go func() {
		fmt.Println("Listen:", p.PublicAddr)
		l, err := net.Listen("tcp", p.PublicAddr)
		if check(err, "tunnel.listen.start") {
			return
		}
		for {
			conn, err := l.Accept()
			check(err, "tunnel.listen.accept")
			go p.forward(conn, p.Tunnel)
		}
	}()
	l, err := net.Listen("tcp", p.TunnelAddr)
	if check(err, "tunnel.listen.start") {
		return
	}
	fmt.Println("Tunnel:", p.TunnelAddr)
	for {
		conn, err := l.Accept()
		if check(err, "tunnel.listen.accept") {
			break
		}
		if conn, ok := conn.(*net.TCPConn); ok {
			check(conn.SetKeepAlive(true))
		}
		if p.Tunnel != nil {
			_ = p.Tunnel.Close()
		}
		p.Tunnel = conn
	}
	return
}
