package main

import (
	"fmt"
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
		Action("server :proxy", server).
		Run()
}

func proxy(c *Context) {
	p := NewProxy(c.GetInt("public"), c.GetInt("tunnel"))
	p.Start()
}

func server(c *Context) {
	s := NewServer("proxy")
	s.Connect()
}
