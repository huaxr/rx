// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/huaxr/rx/util"
)

type RspCtxI interface {
	// JSON response
	JSON(status int16, response interface{})
}

type responseContext struct {
	body   bytes.Buffer
	status int16

	rspHeaders map[string]interface{}
	rspBody    []byte
	time       time.Time
}

func (r *responseContext) getRspStatus() int16 {
	return r.status
}

func (res *responseContext) wrapResponse() {
	res.wrapResponseHeader()
	if len(res.rspBody) > 0 {
		res.body.Write([]byte(fmt.Sprintf("Content-Length: %d", len(string(res.rspBody))) + "\r\n"))
		res.wrapResponseBody()
	}
	res.rspBody = []byte{}
}

func (res *responseContext) wrapResponseHeader() {
	firstLine := fmt.Sprintf("HTTP/1.1 %d\r\n", res.status)
	res.body.Write([]byte(firstLine))
	for k, v := range res.rspHeaders {
		res.body.Write([]byte(k + ": " + v.(string) + "\r\n"))
	}
}

func (res *responseContext) wrapResponseBody() {
	res.body.Write([]byte("\r\n"))
	res.body.Write(res.rspBody)
}

func (res *responseContext) rspToBytes() []byte {
	res.wrapResponse()
	return res.body.Bytes()
}

func (rsp *responseContext) JSON(status int16, response interface{}) {
	rsp.status = status
	rsp.rspHeaders["Content-Type"] = util.MIMEJSON
	bits, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}
	rsp.rspBody = append(rsp.rspBody, bits...)
}

func (rsp *responseContext) stopTime() time.Time {
	return rsp.time
}

func (rsp *responseContext) SetStopTime(t time.Time) {
	rsp.time = t
}
