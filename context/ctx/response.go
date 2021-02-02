// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/huaxr/rx/internal"
	"log"
	"time"
)

type RspCtxI interface {
	GetStatus() int16
	// Bytes return the context response byte
	RspToBytes() []byte
	JSON(status int16, response interface{})
	Abort(status int16, message string)
}

type ResponseContext struct {
	// body all the bytes to response
	body   bytes.Buffer
	status int16

	rspHeaders map[string]interface{}
	rspBody   []byte
	// time response time
	time time.Time
}

func (r *ResponseContext) GetStatus() int16 {
	return r.status
}

func (res *ResponseContext) wrapResponse() {
	res.wrapResponseHeader()
	if len(res.rspBody) > 0 {
		res.body.Write([]byte(fmt.Sprintf("Content-Length: %d", len(string(res.rspBody))) + "\r\n"))
		res.wrapResponseBody()
	}
}

func (res *ResponseContext) wrapResponseHeader() {
	firstLine := fmt.Sprintf("HTTP/2 %d\r\n", res.status)
	res.body.Write([]byte(firstLine))
	for k, v := range res.rspHeaders {
		res.body.Write([]byte(k + ": " + v.(string) + "\r\n"))
	}
}

func (res *ResponseContext) wrapResponseBody() {
	res.body.Write([]byte("\r\n"))
	res.body.Write(res.rspBody)
}

func (res *ResponseContext) RspToBytes() []byte {
	defer res.body.Reset()
	res.wrapResponse()
	return res.body.Bytes()
}

func (rsp *ResponseContext) JSON(status int16, response interface{}) {
	rsp.status = status
	rsp.rspHeaders["Content-Type"] = internal.MIMEJSON
	bits, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}
	rsp.rspBody = bits
}

func (rsp *ResponseContext) Abort(status int16, message string){
	rsp.status = status
	rsp.rspHeaders["Content-Type"] = internal.MIMEHTML
	rsp.rspBody = []byte(message)
}