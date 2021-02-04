# RX Web Framework

RX is a dynamic configurable and assemble web framework based on Go.


## Contents
- [Dynamic middleware](#Dynamic-middleware)
- [Customized Group](#Customized-Group)
- [Customized Response](#Customized-Response)
- [Customized Abort](#Customized-abort)
- [Context Transfer](#Context-Transfer)
- [Epoll Kqueue](#Epoll-Kqueue)
- [URL Parse](#URL-Parse)
- [Body Parse](#Body-Parse)
---

## Dynamic middleware
动态中间件(ctx 栈调用)
使用 Next 函数可以定义本 handler 执行完毕后下一个运行的 handler，动态化中间件，
避免了中间件的静态硬编码带来的逻辑麻烦.

```go
func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler")
	ctx.Next(handler4)
}
```
---

## Customized Group
- 自定义路由组
```go
group := ctx.Group("/api/auth/", handler4, handler2)
group.Register("get", "user", handler)
group.Register("get", "user2", handler)
```
---

## Customized Response
- 自定义 response
- 目前支持JSON
```go
func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler")
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": ctx.Get("user")})
}
```
---

## Customized Abort
- 自定义 abort
- 当在handler context中调用此函数， 将直接以指定的 message 退出当前请求， 后续的栈中中间件将不再执行。如下所示，Next将handler压栈，但是不会再执行
- 富化中 增加模板
```go
ctx.SetDefaultHandler(404, func(ctx ctx.ReqCxtI) {
    ctx.Next(handler) // not execute
    ctx.Abort(200, "Not Allowed")
})
```
---

## Context Transfer
- context 信息传递
- 原生并发处理
```go
ctx.Set("user", []string{"a", "b", "c"})
ctx.Get("user")
```
----

## Epoll&Kqueue
- epoll kqueue 内核态长连接支持
```go
type PollIF interface {
	ChangeRW(fd int)
	ChangeDetach(fd int)
	ChangeRead(fd int)
	AddRW(fd int)
	AddRead(fd int)
	Looping(execute func(fd int) error)
	Close() error
	Trigger(note interface{}) error
}
```
---

## URL Parse
- 获取url参数
```go
name := ctx.GetQuery("name", "XR")
log.Println(name)
```

---

## Body Parse
- post Body动态解析
```go
type PostBody struct {
	Name string `json:"name"`
}

func handler5(ctx ctx.ReqCxtI) {
	var post PostBody
	err := ctx.ParseBody(&post)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": post.Name})
}
```

---