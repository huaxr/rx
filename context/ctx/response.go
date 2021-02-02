// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"fmt"
)

type RspCtxI interface {
	// Bytes return the context response byte
	Bytes() []byte
	GetResponseStatus() int16
	GetResponse() *ResponseContext
}

type ResponseContext struct {
	// body all the bytes to response
	body   bytes.Buffer
	status int16

	rspHeaders map[string]string
	rspBody   []byte
}

func (r *ResponseContext) GetResponse() *ResponseContext {
	return r
}

func (r *ResponseContext) GetResponseStatus() int16 {
	return r.status
}

func (res *ResponseContext) wrapResponse() {
	res.wrapResponseHeader()
	if len(res.rspBody) > 0 {
		res.body.Write([]byte(fmt.Sprintf("Content-Length: %d", len(string(res.rspBody))) + "\r\n"))
		res.wrapResponseBody()
	}
	//log.Println(string(res.body.Bytes()))
}

func (res *ResponseContext) wrapResponseHeader() {
	firstLine := fmt.Sprintf("HTTP/2 %d\r\n", res.status)
	res.body.Write([]byte(firstLine))
	for k, v := range res.rspHeaders {
		res.body.Write([]byte(k + ": " + v + "\r\n"))
	}
}

func (res *ResponseContext) wrapResponseBody() {
	res.body.Write([]byte("\r\n"))
	switch res.status {
	case 200:
		res.body.Write(res.rspBody)
	case 404:
		res.body.Write([]byte("Page not found"))
	case 500:
		res.body.Write([]byte("Internal server error"))
	default:

	}
}

func (res *ResponseContext) Bytes() []byte {
	defer res.body.Reset()
	res.wrapResponse()
	return res.body.Bytes()
}
