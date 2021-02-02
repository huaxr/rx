package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huaxr/rx/context"
	"github.com/huaxr/rx/context/ctx"
	"time"
)

func handler(ctx ctx.ReqCxtI) {
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "context": ctx.Get("user")})
}

func handler3(ctx ctx.ReqCxtI) {
	ctx.Set("a", []string{"a", "b", "c"})
}

func handler2(ctx ctx.ReqCxtI) {
	ctx.Set("user", []string{"a", "b", "c"})
	ctx.Push(handler3)
}

// std epoll
func main() {
	server := context.NewServer("std", "127.0.0.1:9999")

	ctx.SetDefaultHandler(404, func (ctx ctx.ReqCxtI){
		ctx.Abort(404, "不存在该地址")
	})
	server.Register("get", "/aaa", handler2, handler)

	server.Run()
}

func main2() {
	r := gin.Default()
	r.Use()

	r.GET("/ping", func(c *gin.Context) {
		c.Set("a", "a")
		c.JSON(200, gin.H{
			"message": time.Now(),
		})
	})
	r.Run(":9999") // 监听并在 0.0.0.0:8080 上启动服务
}
