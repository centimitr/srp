package main

import (
	"github.com/gorilla/websocket"
	"net"
)

type Server struct {
	ProxyAddr   string
	ServiceAddr string
	Tunnel      *websocket.Conn
	Conn        *net.TCPConn
	Listener    net.Listener
}

func NewServer(proxyAddr, serviceAddr string) *Server {
	return &Server{ProxyAddr: proxyAddr, ServiceAddr: serviceAddr}
}

func (s *Server) CreateService() (err error) {
	s.Listener, err = net.Listen("tcp", s.ServiceAddr)
	if err != nil {
		return
	}
	s.Conn, err = DialTCPKeepAlive(s.ServiceAddr)
	return
}

func (s *Server) ConnectTunnel() (err error) {
	d := new(websocket.Dialer)
	s.Tunnel, _, err = d.Dial(s.ProxyAddr, nil)
	if err != nil {
		return
	}
	log("connect:", s.Tunnel.RemoteAddr())
	go func() {
		for {
			_, req, err := s.Tunnel.ReadMessage()
			if check(err) {
				break
			}
			if s.Conn == nil {
				break
			}
			_, err = s.Conn.Write(req)
			if check(err) {
				break
			}
			resp, err := ReadHTTPMessage(s.Conn)
			if check(err) {
				break
			}
			err = s.Tunnel.WriteMessage(websocket.TextMessage, resp)
			if check(err) {
				break
			}
		}
	}()
	return
}
