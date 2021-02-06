package main

import (
	"github.com/huaxr/rx/logger"
	"runtime"
	"time"

	"github.com/huaxr/rx/ctx"
)

func handler1(c ctx.ReqCxtI) {
	//c.SetDefaultTimeOut(1 * time.Second)
	//c.SetDefaultTTL(4)
	logger.Log.Info("execute handler1", runtime.NumGoroutine())
	c.RegisterStrategy(&ctx.StrategyContext{Ttl: 4,  Async:true})
	//c.SetTimeOut(100 * time.Millisecond)
}

func handler2(c ctx.ReqCxtI) {
	// 使用异步来标识异步处理逻辑
	time.Sleep(2 * time.Second)
	logger.Log.Info("execute handler2")
}

func handler3(c ctx.ReqCxtI) {
	logger.Log.Info("execute handler3")
}

type PostBody struct {
	Name string `json:"name"`
}

func last(c ctx.ReqCxtI) {
	var post PostBody
	err := c.ParseBody(&post)
	c.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": post.Name})
	logger.Log.Info("execute last", err, post.Name)
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

// std epoll
func main() {

	logger.Log.Critical("aaa", )
	logger.Log.Error("bbb", )
	logger.Log.Warning("ccc", )
	logger.Log.Info("ddd %v", "sss")
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

	ctx.Register("post", "/ccc", handler1, handler2, handler5)
	ctx.Register("get", "/ccc", handler1, handler2, handler4)
}
