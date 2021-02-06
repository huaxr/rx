// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/huaxr/rx/internal"
	"github.com/huaxr/rx/logger"

	"go.uber.org/atomic"
)

type ReqCxtI interface {
	RspCtxI
	AbortI
	StrategyI
	// GetHeaders
	GetHeaders() map[string]interface{}
	GetMethod() string
	GetClientAddr() string
	// GetBody return the request body bytes
	GetBody() []byte
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
}

type mod int

const (
	Std mod = iota
	EPoll
)

// RequestContext contains all the ctx when a req has triggered
type RequestContext struct {
	*responseContext
	// abort preStatus will not execute or generate the response handler
	*abortContext
	*StrategyContext

	conn net.Conn
	mod  mod
	// method request method
	method string
	// path request location
	path string
	// clientAddr the remote addr
	clientAddr net.Addr
	// time request time time.now()
	time time.Time
	// headers request header key-value
	reqHeaders map[string]interface{}
	// reqParameters the request raw query parse to map
	reqParameters url.Values
	// body request body bytes
	reqBody bytes.Buffer

	flashStore *sync.Map

	// abort will set the it true
	finished *atomic.Bool
	// stack record the executable func
	stack *stack
	// whether open strategyContext in this context
	// when calling RegisterStrategy, this flag will
	// turn to true.
	openStrategy atomic.Bool
}

var reqCtxPool = sync.Pool{
	New: func() interface{} {
		return &RequestContext{
			responseContext: &responseContext{
				rspHeaders: make(map[string]interface{}),
				rspBody:    []byte{},
				body:       bytes.Buffer{},
			},
			reqHeaders: make(map[string]interface{}),
		}
	},
}

func putContext(rc *RequestContext) {
	rc.free()
	reqCtxPool.Put(rc)
}

func (r *RequestContext) getPath(real bool) string {
	if real {
		return r.path
	}
	return strings.ToLower(r.method) + "::" + r.path
}

func (r *RequestContext) GetHeaders() map[string]interface{} {
	return r.reqHeaders
}

func (r *RequestContext) GetMethod() string {
	return r.method
}

func (r *RequestContext) GetBody() []byte {
	return r.reqBody.Bytes()
}

func (r *RequestContext) GetClientAddr() string {
	if r.clientAddr == nil {
		return ""
	}
	return r.clientAddr.String()
}

// SetClientAddr set the address for logger usage
func (r *RequestContext) setClientAddr(addr net.Addr) {
	r.clientAddr = addr
}

func (r *RequestContext) setAbort(status int16, message interface{}) {
	r.status = status
	switch message.(type) {
	default:
		r.responseContext.rspHeaders["Content-Type"] = internal.MIMEHTML
	case map[string]interface{}:
		r.responseContext.rspHeaders["Content-Type"] = internal.MIMEJSON
	}
	r.abortContext = r.NewAbort(status, message)
	r.finish()
}

func (r *RequestContext) isAbort() bool {
	return r.abortContext != nil || r.finished.Load()
}

// Free when get ReqCxtI from sync.pool, clear the allocated map fields.
func (r *RequestContext) free() *RequestContext {
	r.abortContext = nil
	//r.responseContext = nil
	//r.StrategyContext = nil
	// if the openStrategy did not set false, it will trigger nil pointer
	// at next Get from the sync.Pool, because the flag not clear automatically when put to sync.Pool.
	r.reqHeaders = map[string]interface{}{}
	r.rspHeaders = map[string]interface{}{}
	r.flashStore = &sync.Map{}
	r.openStrategy.Store(false)
	r.finished.Store(false)
	//r.StrategyContext = openDefaultStrategy()
	return r
}

func wrapRequest(buffer []byte) *RequestContext {
	var err error
	reqCtx := reqCtxPool.Get().(*RequestContext)
	reqCtx.time = time.Now()
	reqCtx.finished = atomic.NewBool(false)
	reqCtx.flashStore = new(sync.Map)

	headerAndBody := bytes.Split(buffer, []byte("\r\n\r\n"))
	if len(headerAndBody) != 2 {
		goto ABORT
	}
	err = reqCtx.wrapRequestHeader(headerAndBody[0])
	if err != nil {
		goto ABORT
	}

	if len(headerAndBody[1]) > 0 {
		err = reqCtx.wrapRequestBody(headerAndBody[1])
		if err != nil {
			goto ABORT
		}
	}
	return reqCtx

ABORT:
	reqCtx.setAbort(403, "Bad Request")
	return reqCtx
}

func (rc *RequestContext) wrapRequestHeader(header []byte) error {
	headerLines := bytes.Split(header, []byte("\r\n"))
	if len(headerLines) == 0 {
		return errors.New("headerLines invalid")
	}

	methodAndPath := string(headerLines[0])
	firstLine := strings.Split(methodAndPath, " ")
	if len(firstLine) != 3 {
		return errors.New("err first line parse")
	}
	queryUrl, err := url.Parse(firstLine[1])
	if err != nil {
		logger.Log.Error("err when handler the query path")
		return err
	}
	// delete version
	params, err := url.ParseQuery(queryUrl.RawQuery)
	if err != nil {
		logger.Log.Error("err when handler the query raw path")
		return err
	}
	rc.reqParameters = params
	rc.method, rc.path = firstLine[0], queryUrl.Path

	for _, line := range headerLines[1:] {
		pair := strings.SplitN(string(line), ":", 2)
		if len(pair) != 2 {
			continue
		}
		// using ToLower in order to avoid that when use ab test.
		// the request header is not formatted like usually.
		// Context-type/Content-Type ex.
		rc.reqHeaders[strings.ToLower(pair[0])] = strings.TrimSpace(pair[1])
	}
	return nil
}

func (rc *RequestContext) wrapRequestBody(body []byte) error {
	var buffer bytes.Buffer
	_, err := buffer.Write(body)
	if err != nil {
		return err
	}
	rc.reqBody = buffer
	return nil
}

func (rc *RequestContext) finish() {
	if !rc.finished.Load() {
		rc.SetStopTime(time.Now())
		rc.finished.Store(true)
		logger.ReqLog(&internal.RequestLogger{
			StartTime: rc.time,
			StopTime: rc.responseContext.time,
			Ip: rc.GetClientAddr(),
			Method: rc.GetMethod(),
			Path: rc.getPath(true),
			Status: rc.status,
		})
	}
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
	}()

	// response data received
	for !rc.isAbort() && rc.stack.length > 0 {
		// demanding processing should be using handlerFunc() to return
		// a chan bool to notify whether this stack has been down.
		rc.stack.Pop()(rc)

		if len(rc.rspBody) > 0 && rc.status != 0 {
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

// the core handler invoke and trigger some events in each handler.
// execute each HandlerFunc on stack which is a dynamically
// choice for your logic flow. stack HandlerFunc can using
// the dynamic push option to return a HandlerFunc to the
// stack peek and execute it when next pop.
func (rc *RequestContext) execute() (response *responseContext){
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Critical(string(internal.PrintStack()))
		}
		if !rc.Async {
			rc.checkAbort()
			rc.finish()
			response = rc.responseContext
			rc.connSend()
		}
	}()

	rc.initStack()
	// if free not set abortContext nil, cached abort status will
	// terminate this abnormal state.
	for !rc.isAbort() && rc.stack.length > 0 {
		if !rc.openStrategy.Load() {
			// no strategy.
			rc.stack.Pop()(rc)
		} else {
			// using timeout. using async.
			if rc.timeout {
				// done channel with buffer, attention here.
				// if no buffer here, some goroutines will
				// deadly block in the end.
				done := make(chan bool, 1)
				go rc.asyncExecute(done)
				select {
				case <- done:
					close(done)
				case <- rc.timeoutSignal:
					rc.handleTimeOut(rc)
				}
				return
			}

			if rc.Async {
				done := make(chan bool)
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
	handles = copyStack(rc.getPath(false))
	if handles == nil {
		handles = newStack()
		handles.Push(defaultHANDLERS[404])
	}
	rc.stack = handles
}

func (rc *RequestContext) GetQuery(key, dft string) string {
	if val := rc.reqParameters.Get(key); val != "" {
		return val
	}
	return dft
}

func (rc *RequestContext) ParseBody(dst interface{}) error {
	contentType, ok := rc.reqHeaders["content-type"]
	if !ok {
		contentType = internal.MIMEPlain
	}
	switch contentType {
	case internal.MIMEPlain:

	case internal.MIMEJSON:
		return json.Unmarshal(rc.GetBody(), &dst)

	case internal.MIMEMultipartPOSTForm:

	default:

	}
	return nil
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
	rc.openStrategy.Store(true)
	rc.StrategyContext = strategy
}
