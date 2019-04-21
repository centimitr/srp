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

func (p *Proxy) forwardMessages(client net.Conn) {
	for {
		log(0)
		req, err := ReadHTTPMessage(client)
		if check(err) {
			break
		}
		log(1)
		log(1, string(req))
		err = p.Tunnel.WriteMessage(websocket.TextMessage, req)
		if check(err) {
			break
		}
		log(2)
		_, resp, err := p.Tunnel.ReadMessage()
		if check(err) {
			break
		}
		log(3)
		log(3, string(resp))
		//_, err = fmt.Fprint(client, resp)
		_, err = client.Write(resp)
		if check(err) {
			break
		}
		log(4)
	}
	return
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
			go p.forwardMessages(conn)
		}
	}()

	upgrader := websocket.Upgrader{}
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		p.Tunnel, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		check(err, "upgrade")
		log("connect:", p.Tunnel.RemoteAddr())
	})
	err = r.Run(p.TunnelAddr)
	check(err, "tunnel.service")
	return
}
