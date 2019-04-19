package main

import (
	"bufio"
	"fmt"
	"net"
)

type Server struct {
	ProxyAddr  string
	PublicAddr string
	Tunnel     net.Conn
	Conn       net.Conn
}

func NewServer(addr string) *Server {
	return &Server{ProxyAddr: addr}
}

func (s *Server) call(msg string) (err error) {
	if s.Conn == nil {
		s.Conn, err = net.Dial("tcp", s.PublicAddr)
		if check(err) {
			return
		}
	}
	_, err = fmt.Fprintln(s.Conn, msg)
	check(err)
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
	s.Tunnel, err = net.Dial("tcp", s.ProxyAddr)
	if check(err, "server.connect") {
		return
	}
	go s.handleTunnelMessages()
	fmt.Printf("Connected: %s (%s)\n", s.ProxyAddr, s.Tunnel.RemoteAddr())
	return
}
