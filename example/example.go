package main

import (
	"github.com/huaxr/rx/ctx"
	"log"
	"time"
)

func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler1")
}

func handler2(ctx ctx.ReqCxtI) {
	log.Println("execute handler2")
}

func handler3(ctx ctx.ReqCxtI) {
	log.Println("execute handler3")
	ctx.Set("a", []string{"a", "b", "c"})
	ctx.Next(handler4)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": "xxxxxx"})
}

func handler4(ctx ctx.ReqCxtI) {
	log.Println("execute handler4")
}

type PostBody struct {
	Name string `json:"name"`
}

func handler5(ctx ctx.ReqCxtI) {
	log.Println(string(ctx.GetBody()))
	var post PostBody
	err := ctx.ParseBody(&post)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": post.Name})
	log.Println("execute handler5", err, post.Name)
	ctx.Next(handler)
	//ctx.Abort(200, "sorry")
}

// std epoll
func main() {
	server := ctx.NewServer("std", "127.0.0.1:9999")
	defer server.Run()

	ctx.SetDefaultHandler(404, func(ctx ctx.ReqCxtI) {
		ctx.Abort(404, "not exist")
	})

	group := ctx.Group("/api/auth", handler4, handler2)
	group.Register("get", "/user", handler)
	group.Register("get", "/user2", handler)
	// 递归组
	sub := group.Group("/v2", handler)
	sub.Register("get", "/la", handler)
	sub2 := sub.Group("/v3", handler)
	sub2.Register("get", "/la", handler)


	ctx.Register("get", "/aaa", handler3, handler2, handler)
	ctx.Register("post", "/bbb", handler5)
}
