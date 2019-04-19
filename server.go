package main

import (
	"bufio"
	"fmt"
	"net"
)

type Server struct {
	ProxyAddr   string
	ServiceAddr string
	Tunnel      net.Conn
	Conn        net.Conn
}

func NewServer(proxyAddr, serviceAddr string) *Server {
	return &Server{ProxyAddr: proxyAddr, ServiceAddr: serviceAddr}
}

func (s *Server) call(msg string) (err error) {
	if s.Conn == nil {
		s.Conn, err = DialTCPKeepAlive(s.ServiceAddr)
		if check(err) {
			return
		}
	}
	_, err = fmt.Fprintln(s.Conn, msg)
	check(err)
	return
}

func (s *Server) handleTunnelMessages() {
	for {
		msg, err := bufio.NewReader(s.Tunnel).ReadString('\n')
		if check(err) {
			break
		}
		err = s.call(msg)
		check(err)
	}
	// may need to detect if handling stops
}

func (s *Server) Connect() (err error) {
	s.Tunnel, err = DialTCPKeepAlive(s.ProxyAddr)
	if check(err, "tunnel.connect") {
		return
	}
	go s.handleTunnelMessages()
	fmt.Printf("Connected: %s (%s)\n", s.ProxyAddr, s.Tunnel.RemoteAddr())
	return
}
