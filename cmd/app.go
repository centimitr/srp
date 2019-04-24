package main

import (
	"github.com/devbycm/scli"
	"github.com/devbycm/srp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	new(scli.App).
		Action("proxy :public :tunnel", proxy).
		Action("server :tunnel :service", server).
		Run()
}

func proxy(c *scli.Context) {
	p := srp.NewProxy(c.Get("public"), c.Get("tunnel"))
	check(p.Run(), "proxy")
}

func server(c *scli.Context) {
	tunnelAddr := c.Get("tunnel")
	serviceAddr := c.Get("service")

	s := srp.NewServer(tunnelAddr, serviceAddr)
	check(s.Start(), "proxy")

	r := gin.New()
	r.GET("", func(c *gin.Context) {
		data := c.Query("x")
		c.String(http.StatusOK, data)
		log(data)
	})
	log("listen:", s.Listener.Addr())
	check(http.Serve(s.Listener, r))
}
