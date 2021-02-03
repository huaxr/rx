// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"errors"
	"go.uber.org/atomic"
	"net"
	"strings"
	"sync"
	"time"
)

type ReqCxtI interface {
	RspCtxI

	// GetPath return the request query path.
	GetPath(real bool) string
	// GetHeaders
	GetHeaders() map[string]interface{}
	GetMethod() string
	GetClientAddr() string
	// SetClientAddr set the address for logger usage
	SetClientAddr(addr net.Addr)
	// Execute execute each HandlerFunc on stack or slice.
	// stack HandlerFunc can using the dynamic push option
	// to return a HandlerFunc to the stack peek and execute
	// it when next pop.
	Execute() RspCtxI

	// SetAbort can make the current flow stop.
	SetAbort(status int16, message string)

	// Free when get ReqCxtI from sync.pool, clear the allocated map fields.
	Free() ReqCxtI

	// StartTime return the time when one ReqCxtI generated.
	StartTime() time.Time
	// Finished implements that the context was executed already.
	Finished() bool

	// Set into sync.map
	Set(key string, val interface{})
	// Get from sync.map
	Get(key string) interface{}
	// Push calling push when execute a HandlerFunc and push the next Handler
	// to execute or jump to anther HandlerFunc to deal with the request.
	Next(handlerFunc HandlerFunc)
}

type AbortContext struct {
	abortStatus int16
	message string
}

// RequestContext contains all the context when a req has triggered
type RequestContext struct {
	//context.Context
	ResponseContext
	// method request method
	method  string
	// path request location
	path    string
	// clientAddr the remote addr
	clientAddr net.Addr
	// time request time time.now()
	time time.Time
	// preStatus will not execute or generate the response handler
	abort *AbortContext
	// headers request header key-value
	reqHeaders map[string]interface{}
	// body request body bytes
	reqBody    bytes.Buffer

	flashStore *sync.Map

	finished *atomic.Bool
	// stack record the executable func
	stack *stack
}

func (r *RequestContext) GetPath(real bool) string {
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

func (r *RequestContext) GetClientAddr() string {
	return r.clientAddr.String()
}

func (r *RequestContext) SetClientAddr(addr net.Addr) {
	r.clientAddr = addr
}

func (r *RequestContext) SetAbort(status int16, message string) {
	r.abort = &AbortContext{
		abortStatus: status,
		message: message,
	}
}

func (r *RequestContext) IsAbort() bool {
	return r.abort != nil
}

func (r *RequestContext) Free() ReqCxtI{
	r.abort = nil
	r.reqHeaders = map[string]interface{}{}
	r.rspHeaders = map[string]interface{}{}
	r.flashStore = &sync.Map{}
	return r
}

func WrapRequest(buffer []byte) ReqCxtI {
	var err error
	reqCtx := GetContext().(*RequestContext)
	reqCtx.time = time.Now()
	reqCtx.finished = atomic.NewBool(false)
	reqCtx.flashStore = new(sync.Map)

	headerAndBody := bytes.Split(buffer, []byte("\r\n\r\n"))
	if len(headerAndBody) == 1 {
		err = reqCtx.wrapRequestHeader(headerAndBody[0])
		if err != nil {
			goto ABORT
		}
	} else if len(headerAndBody) == 2 {
		err = reqCtx.wrapRequestHeader(headerAndBody[0])
		if err != nil {
			goto ABORT
		}
		err = reqCtx.wrapRequestBody(headerAndBody[1])
		if err != nil {
			goto ABORT
		}
	} else {
		goto ABORT
	}
	return reqCtx

ABORT:
	reqCtx.SetAbort(403, "Bad Request")
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
	// delete version
	rc.method, rc.path = firstLine[0], firstLine[1]

	for _, line := range headerLines[1:] {
		pair := strings.Split(string(line), ": ")
		if len(pair) != 2 {
			continue
		}
		rc.reqHeaders[pair[0]] = pair[1]
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
	rc.SetStopTime(time.Now())
	rc.finished.Store(true)
	Log(rc)
}

func (rc *RequestContext) Execute() RspCtxI {
	rc.initStack()
	for rc.stack.length > 0 {
		s := rc.stack.Pop()
		s(rc)
	}
	rc.finish()
	return &rc.ResponseContext
}

func (rc *RequestContext) GetDefaultHandlerSlice() []HandlerFunc{
	if rc.IsAbort() {
		handlers := []HandlerFunc{}
		status := rc.abort.abortStatus
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

func (rc *RequestContext) getDefaultHandlerStack() *stack{
	if rc.IsAbort() {
		handlers := NewStack()
		status := rc.abort.abortStatus
		switch {
		case status >= 100 && status <= 199:
			return nil
		default:
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
	val, _ :=  rc.flashStore.Load(key)
	return val
}

func (rc *RequestContext) StartTime() time.Time {
	return rc.time
}

func (rc *RequestContext) Finished() bool {
	return rc.finished.Load()
}

func (rc *RequestContext) Next(handlerFunc HandlerFunc) {
	rc.stack.Push(handlerFunc)
}

func (rc *RequestContext) initStack() {
	handles := rc.getDefaultHandlerStack()
	if handles != nil {
		return
	}
	handles = copyStack(rc.GetPath(false))
	if handles == nil {
		handles = NewStack()
		handles.Push(defaultHANDLERS[404])
	}

	rc.stack = handles
}