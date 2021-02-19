package main

import (
	"log"
	"time"

	"github.com/huaxr/rx/logger"

	"github.com/huaxr/rx/ctx"

	"github.com/huaxr/rx/ctx/engine"

	_ "net/http/pprof"
)

func handler1(c ctx.ReqCxtI) {
	//c.SetDefaultTimeOut(1 * time.Second)
	//c.SetDefaultTTL(4)
	//logger.Log.Info("execute handler1", runtime.NumGoroutine())
	c.RegisterStrategy(
		&ctx.StrategyContext{
			Ttl:   4,
			Async: false,
			//Timeout: 1*time.Second,
		})
}

func handler2(c ctx.ReqCxtI) {
	// 使用异步来标识异步处理逻辑
	//time.Sleep(2 * time.Second)
	//logger.Log.Info("execute handler2")
}

func handler3(c ctx.ReqCxtI) {
	//logger.Log.Info("execute handler3")
}

type PostBody struct {
	Name string `json:"name"`
}

func last(c ctx.ReqCxtI) {

	//time.Sleep(3 * time.Second)
	//var post PostBody
	//err := c.ParseBody(&post)
	c.JSON(200, map[string]interface{}{"time": time.Now()})
	//logger.Log.Info("execute last")
	//engine.Next(handler3)
	//engine.Abort(200, "sorry")
}

func handler4(c ctx.ReqCxtI) {
	//logger.Log.Info("execute handler4")
}

func handler5(c ctx.ReqCxtI) {
	//logger.Log.Info("execute handler5")
	c.JSON(200, "XXX")
}

func upload(c ctx.ReqCxtI) {
	log.Println("zzzzzzzzzzzzzzzzzzz")
	c.SaveLocalFile("/Users/hua/go/src/github.com/huaxr/rx/example/upload")
	c.JSON(200, "xxx")
}

func ping(c ctx.ReqCxtI) {
	c.JSON(200, map[string]string{"message": "pong"})
}

// std epoll
func main() {

	logger.Log.Critical("xxxxxx")

	logger.Log.Error("xxxxxx")
	logger.Log.Warning("xxxxxx")

	server := engine.NewServer("std", "127.0.0.1:9999")
	server.Type("http")
	//server.Type("tcp")
	defer server.Run()

	ctx.SetDefaultHandler(404, func(ctx ctx.ReqCxtI) {
		ctx.Abort(404, "not exist")
	})

	ctx.DisableLog()

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
