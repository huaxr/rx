
# RX 一套基于Go的web框架

### 支持

- ctx 栈调用 
- 自定义 response
- 自定义 abort 
- epoll kqueue 内核态长连接支持
- context 信息传递


```go
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
    // std epoll
	server := context.NewServer("std", "127.0.0.1:9999")
	defer server.Run()

	ctx.SetDefaultHandler(404, func (ctx ctx.ReqCxtI){
		ctx.Abort(404, "not exist")
	})

	//ctx.Group("/api/auth", handler4)

	ctx.Register("get", "/aaa", handler3, handler2, handler)
}
```




