package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net"
)

type Proxy struct {
	PublicAddr string
	TunnelAddr string
	Tunnel     *websocket.Conn
}

func NewProxy(publicAddr, tunnelAddr string) *Proxy {
	return &Proxy{PublicAddr: publicAddr, TunnelAddr: tunnelAddr}
}

func (p *Proxy) forwardMessages(client net.Conn) {
	var req, resp []byte
	for {
		req, _ = ReadHTTPMessage(client)
		if p.Tunnel != nil {
			_ = p.Tunnel.WriteMessage(websocket.TextMessage, req)
			_, resp, _ = p.Tunnel.ReadMessage()
		}
		_, _ = client.Write(resp)
	}
}

func (p *Proxy) ListenPublic() (err error) {
	log("listen:", p.PublicAddr)
	l, err := net.Listen("tcp", p.PublicAddr)
	if check(err, "public") {
		return
	}
	for {
		conn, err := l.Accept()
		if !check(err, "public") {
			go p.forwardMessages(conn)
		}
	}
}

func (p *Proxy) ListenTunnel() (err error) {
	upgrader := websocket.Upgrader{}
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		p.Tunnel, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		if !check(err, "upgrade") {
			log("connect:", p.Tunnel.RemoteAddr())
		}
	})
	err = r.Run(p.TunnelAddr)
	check(err, "tunnel")
	return
}

func (p *Proxy) Run() (err error) {
	go func() {
		_ = p.ListenTunnel()
	}()
	return p.ListenPublic()
}
