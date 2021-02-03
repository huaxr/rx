package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huaxr/rx/context"
	"github.com/huaxr/rx/context/ctx"
	"log"
	"time"
)

func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler1", ctx.Get("a"))
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "context": ctx.Get("user")})
}

func handler2(ctx ctx.ReqCxtI) {
	log.Println("execute handler2")
	ctx.Set("user", []string{"a", "b", "c"})
}

func handler3(ctx ctx.ReqCxtI) {
	log.Println("execute handler3")
	ctx.Set("a", []string{"a", "b", "c"})
	time.Sleep(5 * time.Second)
	ctx.Next(handler4)
}

func handler4(ctx ctx.ReqCxtI) {
	log.Println("execute handler4")
}

// std epoll
func main() {
	server := context.NewServer("std", "127.0.0.1:9999")
	defer server.Run()

	ctx.SetDefaultHandler(404, func (ctx ctx.ReqCxtI){
		ctx.Abort(404, "not exist")
	})

	//ctx.Group("/api/auth", handler4)

	ctx.Register("get", "/aaa", handler3, handler2, handler)
}

func main2() {
	r := gin.Default()
	r.Use()

	r.GET("/ping", func(c *gin.Context) {
		c.Next()
		c.Set("a", "a")
		c.JSON(200, gin.H{
			"message": time.Now(),
		})
	})
	r.Run(":9999") // 监听并在 0.0.0.0:8080 上启动服务
}
