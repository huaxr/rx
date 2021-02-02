// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"time"
)

type ReqCxtI interface {
	//context.Context
	RspCtxI
	GetPath(real bool) string
	GetHeaders() map[string]interface{}
	GetMethod() string
	GetClientAddr() string

	SetClientAddr(addr net.Addr)
	Execute(handlerFunc ...HandlerFunc) RspCtxI

	// SetAbort can make the current flow stop
	SetAbort(status int16, message string)
	GetDefaultHandler() []HandlerFunc
	GetHandlerByStatus(status int16) []HandlerFunc

	Free() ReqCxtI
}

type AbortContext struct {
	abortStatus int16
	message string
}

// RequestContext contains all the context when a req has triggered
type RequestContext struct {
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
	r.rspHeaders = map[string]interface{}{}
	r.rspHeaders = map[string]interface{}{}
	return r
}

func WrapRequest(buffer []byte) ReqCxtI {
	var err error
	reqCtx := GetContext().(*RequestContext)
	headerAndBody := bytes.Split(buffer, []byte("\r\n\r\n"))
	if len(headerAndBody) == 1 {
		err = reqCtx.wrapRequestHeader(headerAndBody[0])
		if err != nil {
			goto ABORT
		}
		return reqCtx
	} else if len(headerAndBody) == 2 {
		err = reqCtx.wrapRequestHeader(headerAndBody[0])
		if err != nil {
			goto ABORT
		}
		err = reqCtx.wrapRequestBody(headerAndBody[1])
		if err != nil {
			goto ABORT
		}
		return reqCtx
	} else {
		goto ABORT
	}

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

func (rc *RequestContext) Execute(handlerFunc ...HandlerFunc) RspCtxI {
	if len(handlerFunc) > 1 {

	}
	handlerFunc(rc)
	return &rc.ResponseContext
}

func (rc *RequestContext) GetDefaultHandler() []HandlerFunc{
	handlers := []HandlerFunc{}
	if rc.IsAbort() {
		status := rc.abort.abortStatus
		switch {
		case status >= 100 && status <= 199:
			return nil
		default:
			handler, ok := HANDLERS[status]
			if !ok {
				handlers = append(handlers, func(ctx ReqCxtI) {
					ctx.Abort(500, STATUS[500])
				})
				return handlers
			}
			handlers = append(handlers, handler)
			return handlers
		}
	}
	return nil
}

func (rc *RequestContext) GetHandlerByStatus(status int16) []HandlerFunc {
	handlers := []HandlerFunc{}
	handlers = append(handlers, HANDLERS[status])
	return handlers
}
