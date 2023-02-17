package api

import (
    "{{ .ProjectName }}/api/{{ .ProjectName }}"
	"{{ .ProjectName }}/api/h"
	"{{ .ProjectName }}/api/proto"

	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(addr string) error {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// generate X-Request-Id
	router.Use(h.MidSetRequestID())

	// logger
	router.Use(h.MidLogger())

	router.Use(h.MidCors()...)

	// local recovery
	router.Use(h.MidRecovery())

	// mount routes
	Mount(&router.RouterGroup)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		// close resource

		os.Exit(1)
	}()

	return router.Run(addr)
}

func Mount(g *gin.RouterGroup) {
	g.GET("metrics", Metrics)
	svc := &{{ .ProjectName}}.Service{}
	proto.RegisterGentestService(g, svc)
}

func Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
