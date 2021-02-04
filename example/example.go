package main

import (
	"github.com/huaxr/rx/ctx"
)

// std epoll
func main() {
	server := ctx.NewServer("std", "127.0.0.1:9999")
	defer server.Run()

	ctx.SetDefaultHandler(404, func(ctx ctx.ReqCxtI) {
		ctx.Abort(404, "not exist")
	})

	group := ctx.Group("/api/auth/", handler4, handler2)
	group.Register("get", "/user", handler)
	group.Register("get", "user2", handler)

	ctx.Register("get", "/aaa", handler3, handler2, handler)
	ctx.Register("post", "/bbb", handler5)
}
