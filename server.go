package main

import (
	"fmt"
	"net"
)

type Server struct {
	//PublicAddr string
	ProxyAddr string
	Conn      net.Conn
	Listener  net.Listener
}

func NewServer(addr string) *Server {
	return &Server{ProxyAddr: addr}
}

func (s *Server) Connect() (err error) {
	s.Listener, err = net.Listen("tcp", ":3001")
	if check(err, "server.connect") {
		return
	}
	conn, err := net.Dial("tcp", s.ProxyAddr)
	if check(err, "server.connect") {
		return
	}
	s.Conn = conn
	fmt.Println("Connected:", s.ProxyAddr)
	return
}
