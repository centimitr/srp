package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"sync"
)

type Proxy struct {
	PublicAddr string
	TunnelAddr string
	Tunnel     *websocket.Conn
}

func NewProxy(publicAddr, tunnelAddr string) *Proxy {
	return &Proxy{PublicAddr: publicAddr, TunnelAddr: tunnelAddr}
}

func copyTCPBuf(from net.Conn, to net.Conn, wg *sync.WaitGroup) {
	_, err := io.Copy(from, to)
	check(err)
	err = from.Close()
	check(err)
	wg.Done()
}

//func (p *Proxy) forward2(from net.Conn, to net.Conn) {
//	var wg sync.WaitGroup
//	wg.Add(2)
//	go copyTCPBuf(to, from, &wg)
//	go copyTCPBuf(from, to, &wg)
//	wg.Wait()
//	fmt.Println("DONE")
//}

func (p *Proxy) forward(client net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		w, err := p.Tunnel.NextWriter(websocket.BinaryMessage)
		check(err, "ws.write")
		_, _ = io.Copy(w, client)
		wg.Done()
	}()
	go func() {
		_, r, err := p.Tunnel.NextReader()
		check(err, "ws.read")
		_, _ = io.Copy(client, r)
		wg.Done()
	}()
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
			go p.forward(conn)
		}
	}()
	//l, err := net.Listen("tcp", p.TunnelAddr)
	//if check(err, "tunnel.listen.start") {
	//	return
	//}

	upgrader := websocket.Upgrader{}
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		p.Tunnel, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		check(err, "upgrade")
	})
	err = r.Run(p.TunnelAddr)
	check(err, "tunnel.service")

	//fmt.Println("Tunnel:", p.TunnelAddr)
	//for {
	//	conn, err := l.Accept()
	//	if check(err, "tunnel.listen.accept") {
	//		break
	//	}
	//	if conn, ok := conn.(*net.TCPConn); ok {
	//		check(conn.SetKeepAlive(true))
	//	}
	//	if p.Tunnel != nil {
	//		_ = p.Tunnel.Close()
	//	}
	//	p.Tunnel = conn
	//}
	return
}
