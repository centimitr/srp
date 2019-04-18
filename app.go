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
		Action("server :tunnel", server).
		Run()
}

func proxy(c *Context) {
	p := c.Get("public")
	t := c.Get("tunnel")
	fmt.Println("Listen:", p)
}

func server(c *Context) {
}

//func proxy(c *Context) {
//	fmt.Println("==", "waiting for registration...")
//	var u *url.URL
//	var rp *httputil.ReverseProxy
//	var l sync.Mutex
//
//	fmt.Println(2)
//	r := gin.New()
//	r.GET("/rp", func(c *gin.Context) {
//		addr := c.Query("addr")
//		u, err := url.Parse(addr)
//		check(err, "addr")
//
//		l.Lock()
//		fmt.Println("=>", addr)
//		rp = httputil.NewSingleHostReverseProxy(u)
//		c.String(http.StatusOK, "=>"+addr)
//		l.Unlock()
//	})
//	r.NoRoute(func(c *gin.Context) {
//		fmt.Println(u)
//		if rp != nil {
//			rp.ServeHTTP(c.Writer, c.Request)
//		}
//	})
//	check(r.Run(c.Get("addr")))
//}
