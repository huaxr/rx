package main

import (
	"log"
	"runtime"
	"time"

	"github.com/huaxr/rx/logger"

	"github.com/gin-gonic/gin"
	"github.com/huaxr/rx/ctx"
)


func handler1(c ctx.ReqCxtI) {
	//c.SetDefaultTimeOut(1 * time.Second)
	//c.SetDefaultTTL(4)
	logger.Log.Info("execute handler1", runtime.NumGoroutine())
	c.RegisterStrategy(
		&ctx.StrategyContext{Ttl: 4,
			Async: true,
			//Timeout: 1*time.Second,
		})
}

func handler2(c ctx.ReqCxtI) {
	// 使用异步来标识异步处理逻辑
	//time.Sleep(2 * time.Second)
	logger.Log.Info("execute handler2")
}

func handler3(c ctx.ReqCxtI) {
	logger.Log.Info("execute handler3")
}

type PostBody struct {
	Name string `json:"name"`
}

func last(c ctx.ReqCxtI) {
	x := c.GetQuery("a", "xx")
	log.Println(x)
	//var post PostBody
	//err := c.ParseBody(&post)
	c.JSON(200, map[string]interface{}{"time": time.Now()})
	logger.Log.Info("execute last")
	//ctx.Next(handler3)
	//ctx.Abort(200, "sorry")
}

func handler4(c ctx.ReqCxtI) {
	logger.Log.Info("execute handler4")
}

func handler5(c ctx.ReqCxtI) {
	logger.Log.Info("execute handler5")
	c.JSON(200, "XXX")
}

func upload(c ctx.ReqCxtI) {
	c.SaveLocalFile("/Users/hua/go/src/github.com/huaxr/rx/example/upload")
	c.JSON(200, "xxx")
}

func ping(c ctx.ReqCxtI) {
	c.JSON(200, "pong")
}

// std epoll
func main() {
	server := ctx.NewServer("std", "127.0.0.1:9999")
	defer server.Run()

	ctx.SetDefaultHandler(404, func(ctx ctx.ReqCxtI) {
		ctx.Abort(404, "not exist")
	})

	group := ctx.Group("/v1", handler1)
	group.Register("get", "ddd", handler2)
	// 递归组
	sub := group.Group("/v2", handler2)
	sub.Register("get", "/aaa", handler3)

	sub2 := sub.Group("/v3", handler3)
	sub2.Register("get", "/eee", last)

	ctx.Register("post", "/ccc", handler5)
	ctx.Register("post", "/upload", upload)

	ctx.Register("get", "/ping", ping)
}

func main2() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func x(c *gin.Context) {
	c.File("/")
	//c.FormFile()

}
