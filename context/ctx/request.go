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
	GetPath() string
	GetHeaders() map[string]interface{}
	GetMethod() string
	GetClientAddr() string

	SetClientAddr(addr net.Addr)
	Execute(handlerFunc HandlerFunc) *ResponseContext
}

// RequestContext contains all the context when a req has triggered
type RequestContext struct {
	// method request method
	method  string
	// path request location
	path    string
	// clientAddr the remote addr
	clientAddr net.Addr
	// time request time time.now()
	time time.Time

	// headers request header key-value
	reqHeaders map[string]interface{}
	// body request body bytes
	reqBody    bytes.Buffer
}

func (r *RequestContext) GetPath() string {
	return strings.ToLower(r.method) + "::" + r.path
}

func (r *RequestContext) GetHeaders() map[string]interface{} {
	return r.reqHeaders
}

func (r *RequestContext) GetMethod() string {
	return r.method
}

func (r *RequestContext) GetLocation() string {
	return r.path
}

func (r *RequestContext) GetClientAddr() string {
	return r.clientAddr.String()
}

func (r *RequestContext) SetClientAddr(addr net.Addr) {
	r.clientAddr = addr
}

func WrapRequest(buffer []byte) (*RequestContext, error) {
	reqCtx := GetContext()
	headerAndBody := bytes.Split(buffer, []byte("\r\n\r\n"))
	if len(headerAndBody) == 1 {
		_ = reqCtx.wrapRequestHeader(headerAndBody[0])
	} else if len(headerAndBody) == 2 {
		_ = reqCtx.wrapRequestHeader(headerAndBody[0])
		_ = reqCtx.wrapRequestBody(headerAndBody[1])
	} else {
		return nil, errors.New("err body format")
	}
	return reqCtx, nil
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

func (rc *RequestContext) Execute(handlerFunc HandlerFunc) *ResponseContext {
	handlerFunc(rc)
	//rc.response.status = 200
	//rc.response.responseBody = resCtx
	//rc.response.header["connection"] = "closed"
	//rc.response.header["content-type"] = "text/html"
	//rc.response.header["content-length"] = "100"
	//rc.response.header["server"] = "GoWeb"
	return nil
}
