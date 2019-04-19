package main

import "net"

func DialTCPKeepAlive(addr string) (c *net.TCPConn, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	c = conn.(*net.TCPConn)
	err = c.SetKeepAlive(true)
	return
}
