# RX Web Framework

RX is a dynamic configurable and assemble web framework based on Go.


## Contents
- [Dynamic middleware](#Dynamic-middleware)
- [Customized Group](#Customized-Group)
- [Customized Strategy](#Customized-Strategy)
- [Customized Response](#Customized-Response)
- [Asynchronous Router](#Asynchronous-Router)
- [Customized Abort](#Customized-abort)
- [TCP Server](#TCP-Server)
- [Context Transfer](#Context-Transfer)
- [Epoll Kqueue](#Epoll-Kqueue)
- [URL Parse](#URL-Parse)
- [Body Parse](#Body-Parse)
- [File Upload](#File-Upload)
- [Benchmark](#Benchmark)
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
  - 异步策略: 设置handler为异步执行， 会往协程池丢送一次任务，并快速返回，任务静默运行。
  - 安全策略: 设置安全检查
  - 熔断策略: 设置熔断检查
  
 todo:
  - panic策略: 
  - error 策略:
  - 兜底策略：
  - 降级策略：
  - 其它:
- 样例1: 如下 group 组注册了 handler1 处理， 其下所有注册的路由都将采取该策略来初始化自己的请求上下文。
- 样例2: 使用 注册接口注册一个自定义的处理策略 c.RegisterStrategy(&ctx.StrategyContext{Ttl: 4})
```go
func handler1(c ctx.ReqCxtI) {
	c.SetTimeOut(1 * time.Second)
	c.SetTTL(4)
	log.Println("execute handler1")
	c.RegisterStrategy(&ctx.StrategyContext{Ttl: 4})
}


group := ctx.Group("/v1", handler1)
// /v1/api 路由若1秒内完成，则退出并返回， 若执行了四次handler，则退出返回
group.Register("get", "api", handler2)
```
---
## Asynchronous Router
- 完美支持异步路由！ 再不用担心写异步逻辑了！ 全部交给RX， 配合自定义策略进行功能化拓展！ 
- 样例:
```go
func handler1(c ctx.ReqCxtI) {
	c.RegisterStrategy(&ctx.StrategyContext{Ttl: 4, Timeout: 2 * time.Second, Async: true})
}

func handler2(c ctx.ReqCxtI) {
	// 使用异步来标识异步处理逻辑
	// 阻塞5秒钟
	time.Sleep(5 * time.Second)
	log.Println("execute handler2")
}
// 执行如下 
group := ctx.Group("/v1", handler1, handler2)
group.Register("get", "ddd", myhandler)

当使用 RegisterStrategy 注册自定义策略绑定到某个Group上， 若在某个handler有延迟逻辑，使用异步可以轻便的让
超时逻辑得以执行！ 
```


---
## Customized Response
- 自定义 response
- 目前支持JSON
```go
func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler")
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "engine": ctx.Get("user")})
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

## TCP Server
- 支持TCP长连接 web
```go
server := engine.NewServer("std", "127.0.0.1:9999")
server.Type("tcp")
defer server.Run()
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
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "engine": post.Name})
}
```

---

## File Upload
- 支持文件上传， 异步上传
- todo: 支持 bytes 流上传
```go
func upload(c ctx.ReqCxtI) {
	c.SaveLocalFile("dst path")
	c.JSON(200, "xxx")
}
```

---

### Benchmark
提升负载策略，真实并发量突破3200
ab -n 1000 -c 10  http://localhost:9999/ping
```
Server Software:        RX
Server Hostname:        127.0.0.1
Server Port:            9999

Document Path:          /ping
Document Length:        18 bytes

Concurrency Level:      10
Time taken for tests:   0.161 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      98000 bytes
HTML transferred:       18000 bytes
Requests per second:    6217.90 [#/sec] (mean)
Time per request:       1.608 [ms] (mean)
Time per request:       0.161 [ms] (mean, across all concurrent requests)
Transfer rate:          595.07 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   0.2      1       1
Processing:     0    1   0.2      1       2
Waiting:        0    1   0.2      1       2
Total:          0    1   0.4      2       2
ERROR: The median and mean for the total time are more than twice the standard
       deviation apart. These results are NOT reliable.

Percentage of the requests served within a certain time (ms)
  50%      2
  66%      2
  75%      2
  80%      2
  90%      2
  95%      2
  98%      2
  99%      2
 100%      2 (longest request)
```