package main

type Server struct {
	ProxyAddr string
}

func NewServer(addr string) *Server {
	return &Server{ProxyAddr: addr}
}

func (s *Server) Connect() {
}
