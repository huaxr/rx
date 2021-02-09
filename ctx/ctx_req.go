// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huaxr/rx/internal"
	"github.com/huaxr/rx/logger"
)

type ReqCxtI interface {
	RspCtxI
	AbortI
	StrategyI

	// Abort response with status
	// setAbort can make the current flow stop.
	// set the stop time & set the finished flag true
	Abort(status int16, message interface{})
	// Set into sync.map
	Set(key string, val interface{})
	// Get from sync.map
	Get(key string) interface{}
	// Push calling push when execute a HandlerFunc and push the next Handler
	// to execute or jump to anther HandlerFunc to deal with the request.
	Next(handlerFunc handlerFunc)

	GetQuery(key, dft string) string
	// ParseBody turn the body bytes to dst
	ParseBody(dst interface{}) error

	// RegisterStrategy register the customized strategy
	RegisterStrategy(strategy *StrategyContext)

	SaveLocalFile(dst string)
}

type mod int

var reg = regexp.MustCompile(`Content-Length: (.*?)\r\n`)

const (
	Std mod = iota
	EPoll
)

// RequestContext contains all the engine when a req has triggered
type RequestContext struct {
	*responseContext
	// abort preStatus will not execute or generate the response handler
	*abortContext
	*StrategyContext

	mod mod
	// raw connection here when mod == Std
	conn    net.Conn
	request *http.Request

	// time request time time.now()
	time       time.Time
	flashStore *sync.Map

	// stack record the executable func
	stack *stack

	// abort will set the it true
	// finished flag represent that the request has done.
	finished bool
}

var reqCtxPool = sync.Pool{
	New: func() interface{} {
		return &RequestContext{
			responseContext: &responseContext{
				rspHeaders: make(map[string]interface{}),
				rspBody:    []byte{},
				body:       bytes.Buffer{},
			},
		}
	},
}

func putContext(rc *RequestContext) {
	rc.free()
	reqCtxPool.Put(rc)
}

// Abort will set the finished flag true
func (r *RequestContext) setAbort(status int16, message interface{}) {
	r.status = status
	switch message.(type) {
	default:
		r.responseContext.rspHeaders["Content-Type"] = internal.MIMEHTML
	case map[string]interface{}:
		r.responseContext.rspHeaders["Content-Type"] = internal.MIMEJSON
	}
	r.abortContext = r.NewAbort(status, message)
	r.finished = true
}

func (r *RequestContext) isAbort() bool {
	return r.abortContext != nil
}

// Free when get ReqCxtI from sync.pool, clear the allocated map fields.
func (r *RequestContext) free() *RequestContext {
	r.abortContext = nil
	//r.responseContext = nil
	//r.StrategyContext = nil
	// if the openStrategy did not set false, it will trigger nil pointer
	// at next Get from the sync.Pool, because the flag not clear automatically when put to sync.Pool.
	r.rspHeaders = map[string]interface{}{}
	r.flashStore = &sync.Map{}
	r.finished = false
	//r.StrategyContext = openDefaultStrategy()
	return r
}

func (r *RequestContext) init() {
	r.time = time.Now()
	r.finished = false
	r.flashStore = new(sync.Map)
}

func WrapStd(conn net.Conn) {
	var buffer bytes.Buffer
	defer buffer.Reset()

	var buf = make([]byte, internal.PieceSize)
	var flag = false
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			// timeout err just break
			break
		}
		buffer.Write(buf[:n])

		if !flag {
			match := reg.FindAllStringSubmatch(buffer.String(), 1)
			// no content-length means the body is small enough
			// to consider read. Or get-method request, just
			// break it.
			if match == nil {
				break
			}
			length, err := strconv.Atoi(match[0][1])
			if err != nil {
				logger.Log.Error(err.Error())
				break
			}
			if length < internal.PieceSize {
				break
			}
			flag = true
		}

		// SetReadDeadline will block read until the deadline
		// for the best efficiency, if the first buf content-length
		// small enough, just break the read option then.
		err = conn.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
		if err != nil {
			logger.Log.Error(err.Error())
			_ = conn.Close()
			return
		}
	}
	buf = buf[:0]

	if buffer.Len() == 0 {
		_ = conn.Close()
		return
	}

	reqCtx := reqCtxPool.Get().(*RequestContext)
	reqCtx.init()

	reqCtx.setMod(Std)
	reqCtx.setRawSock(conn)

	r, err := http.ReadRequest(bufio.NewReader(&buffer))
	if err != nil {
		logger.Log.Error("err: %v, buffer size: %d", err.Error(), buffer.Len())
		reqCtx.setAbort(403, "Bad Request")
	} else {
		reqCtx.setRequest(r)
	}
	reqCtx.execute()
}

func WapEPoll(buf []byte) []byte {
	reqCtx := reqCtxPool.Get().(*RequestContext)
	reqCtx.init()
	reqCtx.setMod(EPoll)
	r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
	reqCtx.request = r
	res := reqCtx.execute()
	return res.rspBody
}

func (rc *RequestContext) setMod(m mod) {
	rc.mod = m
}

func (rc *RequestContext) setRawSock(c net.Conn) {
	rc.conn = c
}

func (rc *RequestContext) setRequest(req *http.Request) {
	rc.request = req
}

func (rc *RequestContext) GetMethod() string {
	if rc.request == nil {
		return ""
	}
	return rc.request.Method
}

func (rc *RequestContext) GetPath() string {
	if rc.request == nil {
		return ""
	}
	return rc.request.URL.Path
}

func (rc *RequestContext) finish() {
	rc.SetStopTime(time.Now())
	rc.finished = true
	logger.ReqLog(&internal.RequestLogger{
		StartTime: rc.time,
		StopTime:  rc.responseContext.time,
		Ip:        rc.conn.RemoteAddr().String(),
		Method:    rc.GetMethod(),
		Path:      rc.GetPath(),
		Status:    rc.status,
	})
	// return back the context poll
	putContext(rc)
}

func (rc *RequestContext) checkAbort() bool {
	if rc.isAbort() {
		rc.responseContext.rspBody = rc.abortContext.GetAbortMessage()
		return true
	}
	return false
}

func (rc *RequestContext) isStd() bool {
	return rc.mod == Std
}

func (rc *RequestContext) isEPoll() bool {
	return rc.mod == EPoll
}

func (rc *RequestContext) setMemorySize(size int64) {
	err := rc.request.ParseMultipartForm(size)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
}

func (rc *RequestContext) GetFileName() string {
	_, header, err := rc.request.FormFile("file")
	if err != nil {
		logger.Log.Error(err.Error())
		return ""
	}
	return header.Filename
}

func (rc *RequestContext) GetFile() multipart.File {
	file, _, err := rc.request.FormFile("file")
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
	return file
}

func (rc *RequestContext) SaveLocalFile(dst string) {
	file, header, err := rc.request.FormFile("file")
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	path := dst + "/" + header.Filename
	cur, err := os.Create(path)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	defer cur.Close()
	_, err = io.Copy(cur, file)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	fi, err := os.Stat(path)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	fmt.Println("file size is ", fi.Size(), err)
}

func (rc *RequestContext) connSend() {
	if !rc.isStd() {
		return
	}
	_, err := rc.conn.Write(rc.responseContext.rspToBytes())
	if err != nil {
		logger.Log.Error(err.Error())
		_ = rc.conn.Close()
	}
	_ = rc.conn.Close()
}

func (rc *RequestContext) asyncExecute(async chan bool) {
	defer func() {
		rc.checkAbort()
		rc.finish()
		rc.connSend()
		close(async)
	}()

	// response data received
	for rc.stack.length.Load() > 0 {
		// demanding processing should be using handlerFunc() to return
		// a chan bool to notify whether this stack has been down.
		h := rc.stack.Pop()
		h(rc)

		if rc.isAbort() || rc.finished {
			// if asyncSignal equals nil, this goroutine
			// will perpetual block eternal die, which will
			// cause memory leak, using runtime.NumGoroutine
			// to debug this.
			async <- true
			return
		}
		// handle ttl here
		if rc.Ttl == 0 {
			rc.handleTTL(rc)
			async <- true
			return
		}
		// stack execute once. ttl -= 1
		rc.decTTL()
	}
}

// the ctx handler invoke and trigger some events in each handler.
// execute each HandlerFunc on stack which is a dynamically
// choice for your logic flow. stack HandlerFunc can using
// the dynamic push option to return a HandlerFunc to the
// stack peek and execute it when next pop.
func (rc *RequestContext) execute() (response *responseContext) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Recovery(string(internal.PrintStack()))
		}
		if rc.StrategyContext != nil && (rc.Async || rc.timeout) {
			return
		}
		rc.checkAbort()
		response = rc.responseContext
		rc.connSend()
		rc.finish()
	}()
	// initStack will set the *stack and abort status.
	rc.initStack()
	// not abort, not finished check with the available stack.
	for !rc.isAbort() && !rc.finished && rc.stack.length.Load() > 0 {
		// not using strategy.
		if rc.StrategyContext == nil {
			rc.stack.Pop()(rc)
		} else {
			//if rc.Demotion != nil {
			//	rc.Demotion.Do()
			//}

			// stop and set abort whether score a hit of the Fusing.Do() method.
			if rc.Fusing != nil && rc.Fusing.Do() {
				rc.setAbort(403, "fusing deny")
				return
			}

			if rc.Security != nil && rc.Security.Do() {
				rc.setAbort(403, "security deny")
				return
			}

			// using timeout. using async, ttl...
			if rc.timeout {
				done := make(chan bool, 1)
				// done channel with buffer, attention here.
				// if no buffer here, some goroutines will
				// deadly block in the end.
				go rc.asyncExecute(done)
				select {
				case <-done:
				case <-rc.timeoutSignal:
					rc.handleTimeOut(rc)
				}
				return
			}

			if rc.Async {
				done := make(chan bool, 1)
				// chan with buffer, otherwise block.
				// todo: using goroutine pool to manager the counts.
				go rc.asyncExecute(done)
				return
			}

			rc.stack.Pop()(rc)
			if rc.Ttl == 0 {
				rc.handleTTL(rc)
				return
			}
			rc.decTTL()
		}
	}
	return
}

func (rc *RequestContext) getDefaultHandlerSlice() []handlerFunc {
	if rc.isAbort() {
		handlers := []handlerFunc{}
		status := rc.abortContext.abortStatus
		switch {
		case status >= 100 && status <= 199:
			return nil
		default:
			handler, ok := defaultHANDLERS[status]
			if !ok {
				handlers = append(handlers, func(ctx ReqCxtI) {
					ctx.Abort(500, defaultSTATUS[500])
				})
				return handlers
			}
			handlers = append(handlers, handler)
			return handlers
		}
	}
	return nil
}

func (rc *RequestContext) getDefaultHandlerStack() *stack {
	if rc.isAbort() {
		status := rc.abortContext.abortStatus
		switch {
		case status >= 100 && status <= 200:
			return nil
		default:
			handlers := newStack()
			handler, ok := defaultHANDLERS[status]
			if !ok {
				handlers.Push(func(ctx ReqCxtI) {
					ctx.Abort(500, defaultSTATUS[500])
				})
				return handlers
			}
			handlers.Push(handler)
			return handlers
		}
	}
	return nil
}

func (rc *RequestContext) Set(key string, val interface{}) {
	rc.flashStore.Store(key, val)
}

func (rc *RequestContext) Get(key string) interface{} {
	val, _ := rc.flashStore.Load(key)
	return val
}

func (rc *RequestContext) startTime() time.Time {
	return rc.time
}

func (rc *RequestContext) Next(handlerFunc handlerFunc) {
	rc.stack.Push(handlerFunc)
}

func (rc *RequestContext) initStack() {
	handles := rc.getDefaultHandlerStack()
	if handles != nil {
		rc.stack = handles
		return
	}
	key := rc.getPathKey()
	handles = copyStack(key)
	if handles == nil {
		handles = newStack()
		handles.Push(defaultHANDLERS[404])
	}
	rc.stack = handles
}

func (rc *RequestContext) getPathKey() string {
	return fmt.Sprintf("%s::", strings.ToLower(rc.request.Method)) + rc.GetPath()
}

func (rc *RequestContext) ParseBody(dst interface{}) error {
	return nil
}

func (rc *RequestContext) GetQuery(key, dft string) string {
	res, ok := rc.request.URL.Query()[key]
	if !ok {
		return dft
	}
	return res[0]
}

func (rc *RequestContext) Abort(status int16, message interface{}) {
	rc.setAbort(status, message)
}

func (rc *RequestContext) RegisterStrategy(strategy *StrategyContext) {
	if strategy == nil {
		strategy = openDefaultStrategy()
	} else {
		strategy.wrapDefault()
	}
	rc.StrategyContext = strategy
}
