# RX Web Framework

RX is a dynamic configurable and assemble web framework based on Go.


## Contents
- [Dynamic middleware](#Dynamic-middleware)
- [Customized Group](#Customized-Group)
- [Customized Strategy](#Customized-Strategy)
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
- 自定义路由组， 可以定义递归组， 每个组可添加多个handler， 以递归栈形式调用。极大简化了手动注册路由的麻烦。
- 例如: 定义一个 /v1 组， 其子组 /v2, 孙组 /v3, 如下可以注册路由为:
```go
group := ctx.Group("/v1", handler1)
group.Register("get", "ddd", handler2)
// 递归组
sub := group.Group("/v2", handler2, handler3)
sub.Register("get", "/aaa", handler3)

sub2 := sub.Group("/v3", handler4, handler5)
sub2.Register("get", "/eee", last)

ctx.Register("post", "/ccc", handler1, handler2, handler4)
ctx.Register("get", "/ccc", handler1, handler2, handler4)

---------------------------------------------------
|EGISTER ROUTER:|   post |                /ccc |3| 
|EGISTER ROUTER:|    get |                /ccc |3| 
|EGISTER ROUTER:|    get |             /v1/ddd |2| 
|EGISTER ROUTER:|    get |          /v1/v2/aaa |4| 
|EGISTER ROUTER:|    get |       /v1/v2/v3/eee |6| 
--------------------------------------------------
```
---

## Customized Strategy
- 自定义策略
  - 超时策略: 在某个handler内部设置超时， 可极大简化reqContext 的超时管理， 无论栈内有多少待执行handler，一旦
  超时直接按照约定的超时策略中断流程， 并返回相关数据 
  - TTL策略: 可设置最大栈内调用深度， 可以极大简化维护栈内调用，方便debug，一旦handler执行次数超过 ttl 设定的阈值，
  将中断reqContext 并按照约定返回相关数据
  - panic策略: 
  - error 策略:
  - 兜底策略：
  - 降级策略：
  - 其它:
- 例如: 如下 group 组注册了 handler1 处理， 其下所有注册的路由都将采取该策略来初始化自己的请求上下文。
```go
func handler1(ctx ctx.ReqCxtI) {
	ctx.SetTimeOut(1 * time.Second)
	ctx.SetTTL(4)
	log.Println("execute handler1")
}

group := ctx.Group("/v1", handler1)
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