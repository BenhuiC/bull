// Code generated by protoc-gen-ginsev. DO NOT EDIT.

package proto

import (
	gin "github.com/gin-gonic/gin"
)

type GentestService interface {
	// 探活
	Ping(c *gin.Context)
}

// RegisterGentestService register router
func RegisterGentestService(router gin.IRouter, service GentestService) {
	router.Handle("GET", "/ping", service.Ping)
}
