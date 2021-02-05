package main

import (
	"log"
	"time"

	"github.com/huaxr/rx/ctx"
)

func handler1(ctx ctx.ReqCxtI) {
	ctx.SetTimeOut(1 * time.Second)
	ctx.SetTTL(4)
	log.Println("execute handler1")
}

func handler2(ctx ctx.ReqCxtI) {
	time.Sleep(999 * time.Millisecond)
	log.Println("execute handler2")
}

func handler3(ctx ctx.ReqCxtI) {
	log.Println("execute handler3")
}

type PostBody struct {
	Name string `json:"name"`
}

func last(ctx ctx.ReqCxtI) {
	var post PostBody
	err := ctx.ParseBody(&post)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": post.Name})
	log.Println("execute last", err, post.Name)
	//ctx.Next(handler3)
	//ctx.Abort(200, "sorry")
}

func handler4(ctx ctx.ReqCxtI) {
	log.Println("execute handler4")
}

func handler5(ctx ctx.ReqCxtI) {
	log.Println("execute handler5")
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
	sub := group.Group("/v2", handler2, handler3)
	sub.Register("get", "/aaa", handler3)

	sub2 := sub.Group("/v3", handler4, handler5)
	sub2.Register("get", "/eee", last)

	ctx.Register("post", "/ccc", handler1, handler2, handler4)
	ctx.Register("get", "/ccc", handler1, handler2, handler4)
}
