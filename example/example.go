package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huaxr/rx/context"
	"github.com/huaxr/rx/context/ctx"
	"time"
)

func handler(ctx *ctx.Context) {
	ctx.JSON(200, map[string]interface{}{"time": time.Now()})
	return
}

// std epoll
func main1() {
	server := context.NewServer("epoll", "127.0.0.1:9999")
	server.Register("get", "/aaa", handler)
	server.Run()
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
