package main

import (
	"fmt"
	"io"
	"net"
	"strings"
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
	wg.Done()
}

func (p *Proxy) forward(from net.Conn, to net.Conn) {
	ip := strings.Split(p.Tunnel.RemoteAddr().String(), ":")[0]
	tunnel, err := net.Dial("tcp", ip+":3001")
	check(err)
	//p.Tunnel = tunnel

	var wg sync.WaitGroup
	wg.Add(2)
	fmt.Println("COPYING")
	go copyTCPBuf(tunnel, from, &wg)
	go copyTCPBuf(from, tunnel, &wg)
	wg.Wait()
	fmt.Println("DONE")
}

func (p *Proxy) Start() (err error) {
	go func() {
		fmt.Println("Listen:", p.PublicAddr)
		l, err := net.Listen("tcp", p.PublicAddr)
		if check(err, "proxy.listen.start") {
			return
		}
		for {
			conn, err := l.Accept()
			check(err, "proxy.listen.accept")
			go p.forward(conn, p.Tunnel)
		}
	}()
	l, err := net.Listen("tcp", p.TunnelAddr)
	if check(err, "proxy.listen.start") {
		return
	}
	fmt.Println("Tunnel:", p.TunnelAddr)
	for {
		conn, err := l.Accept()
		if check(err, "proxy.listen.accept") {
			break
		}
		if p.Tunnel != nil {
			_ = p.Tunnel.Close()
		}
		if conn, ok := conn.(*net.TCPConn); ok {
			check(conn.SetKeepAlive(true))
		}
		p.Tunnel = conn
	}
	return
}
