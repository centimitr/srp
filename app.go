package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func check(err error, prompts ...string) bool {
	if err != nil {
		fmt.Println(strings.Join(prompts, ": ")+":", err)
		return true
	}
	return false
}

func main() {
	new(App).
		Action("proxy :public :tunnel", proxy).
		Action("server :tunnel :service", server).
		Run()
}

func proxy(c *Context) {
	p := NewProxy(c.Get("public"), c.Get("tunnel"))
	check(p.Start())
}

func server(c *Context) {
	s := NewServer(c.Get("tunnel"), c.Get("service"))
	check(s.Connect(), "connect")

	r := gin.New()
	r.GET("", func(c *gin.Context) {
		c.String(http.StatusOK, "hello"+c.Query("x"))
		fmt.Println("hahahah")
	})
	check(http.Serve(s.Listener, r))
	//check(r.Run(s.Conn.LocalAddr().String()), "listen")
}
