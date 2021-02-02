package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huaxr/rx/context"
	"github.com/huaxr/rx/context/ctx"
	"time"
)

func handler(ctx ctx.ReqCxtI) {
	ctx.JSON(200, map[string]interface{}{"time": time.Now()})
}

// std epoll
func main() {
	server := context.NewServer("std", "127.0.0.1:9999")

	server.Register("get", "/aaa", ctx.Log, handler)

	server.Run()
}

func main2() {
	r := gin.Default()
	r.Use()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
