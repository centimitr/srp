package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
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

//func (s *Server) handleTunnelMessages() {
//	var wg sync.WaitGroup
//	wg.Add(2)
//	go copyTCPBuf(s.Tunnel, s.Conn, &wg)
//	go copyTCPBuf(s.Conn, s.Tunnel, &wg)
//	wg.Wait()
//
//	//for {
//	//	//_, _ = io.Copy(s.Conn, s.Tunnel)
//	//	msg, err := bufio.NewReader(s.Tunnel).ReadString('\n')
//	//	if check(err) {
//	//		break
//	//	}
//	//	err = s.call(msg)
//	//	check(err, "call")
//	//}
//	// may need to detect if handling stops
//}

//func (s *Server) ConnectTunnel2() (err error) {
//	s.Tunnel, err = DialTCPKeepAlive(s.ProxyAddr)
//	if err != nil {
//		return
//	}
//	go s.handleTunnelMessages()
//	fmt.Printf("Connected: %s (%s)\n", s.ProxyAddr, s.Tunnel.RemoteAddr())
//	return
//}

func (s *Server) ConnectTunnel() (err error) {
	d := new(websocket.Dialer)
	s.Tunnel, _, err = d.Dial(s.ProxyAddr, nil)
	go func() {
		for {
			for {
				messageType, r, err := s.Tunnel.NextReader()
				check(err)
				_, err = io.Copy(s.Conn, r)
				check(err)

				w, err := s.Tunnel.NextWriter(messageType)
				check(err)
				_, err = io.Copy(w, s.Conn)
				check(err)
				err = w.Close()
				check(err)
			}
		}
	}()
	//s.Tunnel, err = DialTCPKeepAlive(s.ProxyAddr)
	//if err != nil {
	//	return
	//}
	//go s.handleTunnelMessages()
	//fmt.Printf("Connected: %s (%s)\n", s.ProxyAddr, s.Tunnel.RemoteAddr())
	return
}
