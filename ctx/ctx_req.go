// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/huaxr/rx/util"

	"go.uber.org/atomic"
)

type ReqCxtI interface {
	RspCtxI
	// GetHeaders
	GetHeaders() map[string]interface{}
	GetMethod() string
	GetClientAddr() string
	// GetBody return the request body bytes
	GetBody() []byte
	// Abort response with status
	// setAbort can make the current flow stop.
	// set the stop time & set the finished flag true
	Abort(status int16, message string)
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
}

// RequestContext contains all the ctx when a req has triggered
type requestContext struct {
	*responseContext
	// abort preStatus will not execute or generate the response handler
	*abortContext

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
}

func (r *requestContext) getPath(real bool) string {
	if real {
		return r.path
	}
	return strings.ToLower(r.method) + "::" + r.path
}

func (r *requestContext) GetHeaders() map[string]interface{} {
	return r.reqHeaders
}

func (r *requestContext) GetMethod() string {
	return r.method
}

func (r *requestContext) GetBody() []byte {
	return r.reqBody.Bytes()
}

func (r *requestContext) GetClientAddr() string {
	return r.clientAddr.String()
}

// SetClientAddr set the address for logger usage
func (r *requestContext) setClientAddr(addr net.Addr) {
	r.clientAddr = addr
}

func (r *requestContext) setAbort(status int16, message string) {
	r.abortContext = NewAbortContext(status, message)
	r.finish()
}

func (r *requestContext) isAbort() bool {
	return r.abortContext != nil
}

// Free when get ReqCxtI from sync.pool, clear the allocated map fields.
func (r *requestContext) free() *requestContext {
	r.abortContext = nil
	r.reqHeaders = map[string]interface{}{}
	r.rspHeaders = map[string]interface{}{}
	r.flashStore = &sync.Map{}
	return r
}

func wrapRequest(buffer []byte) *requestContext {
	var err error
	reqCtx := GetContext()
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

func (rc *requestContext) wrapRequestHeader(header []byte) error {
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
		log.Println("err when handler the query path")
		return err
	}
	// delete version
	params, err := url.ParseQuery(queryUrl.RawQuery)
	if err != nil {
		log.Println("err when handler the query raw path")
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

func (rc *requestContext) wrapRequestBody(body []byte) error {
	var buffer bytes.Buffer
	_, err := buffer.Write(body)
	if err != nil {
		return err
	}
	rc.reqBody = buffer
	return nil
}

func (rc *requestContext) finish() {
	rc.SetStopTime(time.Now())
	rc.finished.Store(true)
	Log(rc)
}

// execute each HandlerFunc on stack or slice.
// stack HandlerFunc can using the dynamic push option
// to return a HandlerFunc to the stack peek and execute
// it when next pop.
func (rc *requestContext) execute() *responseContext {
	rc.initStack()
	for rc.stack.length > 0 {
		// if the ctx abort by an Abort calling.
		if rc.abortContext != nil || rc.finished.Load() {
			rc.stack = nil
			return rc.responseContext
		}
		s := rc.stack.Pop()
		s(rc)
	}
	rc.finish()
	return rc.responseContext
}

func (rc *requestContext) getDefaultHandlerSlice() []handlerFunc {
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

func (rc *requestContext) getDefaultHandlerStack() *stack {
	if rc.isAbort() {
		status := rc.abortContext.abortStatus
		switch {
		case status >= 100 && status <= 200:
			return nil
		default:
			handlers := NewStack()
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

func (rc *requestContext) Set(key string, val interface{}) {
	rc.flashStore.Store(key, val)
}

func (rc *requestContext) Get(key string) interface{} {
	val, _ := rc.flashStore.Load(key)
	return val
}

func (rc *requestContext) startTime() time.Time {
	return rc.time
}

func (rc *requestContext) Finished() bool {
	return rc.finished.Load()
}

func (rc *requestContext) Next(handlerFunc handlerFunc) {
	rc.stack.Push(handlerFunc)
}

func (rc *requestContext) initStack() {
	handles := rc.getDefaultHandlerStack()
	if handles != nil {
		rc.stack = handles
		return
	}
	handles = copyStack(rc.getPath(false))
	if handles == nil {
		handles = NewStack()
		handles.Push(defaultHANDLERS[404])
	}
	rc.stack = handles
}

func (rc *requestContext) GetQuery(key, dft string) string {
	if val := rc.reqParameters.Get(key); val != "" {
		return val
	}
	return dft
}

func (rc *requestContext) ParseBody(dst interface{}) error {
	contentType, ok := rc.reqHeaders["content-type"]
	if !ok {
		contentType = util.MIMEPlain
	}
	switch contentType {
	case util.MIMEPlain:

	case util.MIMEJSON:
		return json.Unmarshal(rc.GetBody(), &dst)

	case util.MIMEMultipartPOSTForm:

	default:

	}
	return nil
}

func (rsp *requestContext) Abort(status int16, message string) {
	rsp.status = status
	rsp.rspHeaders["Content-Type"] = util.MIMEHTML
	// replace any bytes which already set.
	rsp.rspBody = []byte(message)
	rsp.setAbort(status, message)
}
