package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"strconv"
	"strings"
	"unicode"
)

func log(vs ...interface{}) {
	fmt.Println(vs...)
}

func check(err error, prompts ...string) bool {
	if err != nil {
		prompt := strings.Join(prompts, ": ")
		if prompt != "" {
			prompt += ":"
			fmt.Println(prompt, err)
		}
		fmt.Println(err)
		return true
	}
	return false
}

func DialTCPKeepAlive(addr string) (c *net.TCPConn, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	c = conn.(*net.TCPConn)
	err = c.SetKeepAlive(true)
	return
}

func isNotNumber(r rune) bool {
	return !unicode.IsNumber(r)
}

func ReadHTTPMessage(conn io.Reader) ([]byte, error) {
	r := bufio.NewReader(conn)
	buf := new(bytes.Buffer)
	var count int
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		buf.Write(line)
		if bytes.HasPrefix(line, []byte("Content-Length:")) {
			count, _ = strconv.Atoi(strings.TrimFunc(string(line), isNotNumber))
		}
		if bytes.Equal(line, []byte("\r\n")) {
			break
		}
	}
	if count != 0 {
		body := make([]byte, count)
		_, err := r.Read(body)
		if err != nil {
			return nil, err
		}
		buf.Write(body)
	}
	return buf.Bytes(), nil
}

func ReadWriteWebsocket(conn *websocket.Conn, messageType int, req []byte) (resp []byte, err error) {
	err = conn.WriteMessage(messageType, req)
	if err != nil {
		return
	}
	_, resp, err = conn.ReadMessage()
	return
}
