package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
)

type Server struct {
	ProxyAddr   string
	ServiceAddr string
	//Tunnel      net.Conn
	Tunnel   *websocket.Conn
	Conn     net.Conn
	Listener net.Listener
}

func NewServer(proxyAddr, serviceAddr string) *Server {
	return &Server{ProxyAddr: proxyAddr, ServiceAddr: serviceAddr}
}

func (s *Server) call(msg string) (err error) {
	fmt.Println("CALL:", msg)
	s.Conn, err = DialTCPKeepAlive(s.ServiceAddr)
	check(err)
	_, err = fmt.Fprintln(s.Conn, msg)
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

func (s *Server) ConnectTunnel() (err error) {
	d := new(websocket.Dialer)
	s.Tunnel, _, err = d.Dial(s.ProxyAddr, nil)
	if err != nil {
		return
	}
	log("connect:", s.Tunnel.RemoteAddr())
	go func() {
		for {
			log(0)
			_, req, err := s.Tunnel.ReadMessage()
			if check(err) {
				break
			}
			log(1)
			log(1, string(req))
			_, err = s.Conn.Write(req)
			if check(err) {
				break
			}
			log(2)
			resp, err := ReadHTTPMessage(s.Conn)
			if check(err) {
				break
			}
			log(3)
			log(3, string(resp))
			err = s.Tunnel.WriteMessage(websocket.TextMessage, resp)
			if check(err) {
				break
			}
			log(4)
		}
	}()
	return
}
