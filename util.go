package main

import (
	"bufio"
	"bytes"
	"fmt"
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
		fmt.Println(strings.Join(prompts, ": ")+":", err)
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

//func copyLine(r *bufio.Reader, buf *bytes.Buffer) (err error) {
//	line, err := r.ReadBytes('\n')
//	if err != nil {
//		return
//	}
//	buf.Write(line)
//	return
//}

func ReadHTTPMessage(conn io.Reader) ([]byte, error) {
	r := bufio.NewReader(conn)
	buf := bytes.NewBuffer([]byte{})
	//readingBody := false
	var count int
	for {
		//copyLine(line)
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
