package srp

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type Server struct {
	ProxyAddr   string
	ServiceAddr string

	Tunnel   *websocket.Conn
	Listener net.Listener
	Conn     *net.TCPConn

	TunnelConnRetryTimes    int
	TunnelConnRetryInterval time.Duration
}

func NewServer(proxyAddr, serviceAddr string) *Server {
	return &Server{
		ProxyAddr:               proxyAddr,
		ServiceAddr:             serviceAddr,
		TunnelConnRetryTimes:    20,
		TunnelConnRetryInterval: 3 * time.Second,
	}
}

func (s *Server) forwardMessages() {
	for {
		if s.Tunnel == nil {
			break
		}
		_, req, err := s.Tunnel.ReadMessage()
		if check(err) {
			break
		}
		if s.Conn == nil {
			continue
		}
		_, err = s.Conn.Write(req)
		if check(err) {
			continue
		}
		resp, err := ReadHTTPMessage(s.Conn)
		if check(err) {
			continue
		}
		err = s.Tunnel.WriteMessage(websocket.TextMessage, resp)
		if check(err) {
			break
		}
	}
	check(s.ConnectTunnel())
}

func (s *Server) ConnectTunnel() (err error) {
	d := new(websocket.Dialer)
	for r := 1; r <= s.TunnelConnRetryTimes; r++ {
		s.Tunnel, _, err = d.Dial(s.ProxyAddr, nil)
		if err == nil {
			log("connect:", s.Tunnel.RemoteAddr())
			go s.forwardMessages()
			break
		}
		if r != s.TunnelConnRetryTimes {
			log("retry: connect:", s.ProxyAddr)
			time.Sleep(s.TunnelConnRetryInterval)
		}
	}
	return
}

func (s *Server) CreateService() (err error) {
	s.Listener, err = net.Listen("tcp", s.ServiceAddr)
	if err != nil {
		return
	}
	s.Conn, err = DialTCPKeepAlive(s.ServiceAddr)
	return
}

func (s *Server) Start() (err error) {
	err = s.CreateService()
	if err != nil {
		return
	}
	err = s.ConnectTunnel()
	return
}
