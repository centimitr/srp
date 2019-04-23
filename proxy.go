package srp

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

func (p *Proxy) disconnectTunnel(err error) {
	if websocket.IsUnexpectedCloseError(err) {
		if p.Tunnel != nil {
			log("disconnect:", p.Tunnel.RemoteAddr())
		}
		p.Tunnel = nil
	}
}

func (p *Proxy) forwardMessages(client net.Conn) {
	var req, resp []byte
	var err error
	for {
		req, err = ReadHTTPMessage(client)
		if check(err, "forward") {
			break
		}
		if p.Tunnel == nil {
			break
		}
		resp, err = ReadWriteWebsocket(p.Tunnel, websocket.TextMessage, req)
		if check(err, "forward") {
			p.disconnectTunnel(err)
			break
		}
		_, err = client.Write(resp)
		check(err, "forward")
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
		if check(err, "public") {
			break
		}
		go p.forwardMessages(conn)
	}
	return
}

func (p *Proxy) ListenTunnel() (err error) {
	upgrader := websocket.Upgrader{}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		p.Tunnel, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		if !check(err, "upgrade") {
			log("connect:", p.Tunnel.RemoteAddr())
		}
	})
	log("tunnel:", p.TunnelAddr)
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
