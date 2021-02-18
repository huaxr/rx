// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huaxr/rx/logger"

	"github.com/huaxr/rx/internal"
)

type RspCtxI interface {
	// JSON response
	JSON(status int16, response interface{})
}

type responseContext struct {
	body bytes.Buffer

	rspHeaders map[string]interface{}
	rspBody    []byte
	status     int16

	time time.Time
}

func (r *responseContext) getRspStatus() int16 {
	return r.status
}

func (res *responseContext) wrapResponse() []byte {
	defer func() {
		res.rspBody = []byte{}
		res.body.Reset()
	}()
	res.wrap()
	return res.body.Bytes()
}

func (res *responseContext) wrap() {
	firstLine := fmt.Sprintf("HTTP/1.1 %d\r\n", res.status)
	res.body.Write(internal.StringToBytes(firstLine))
	for k, v := range res.rspHeaders {
		res.body.Write(internal.StringToBytes(k + ": " + v.(string) + "\r\n"))
	}

	res.body.Write(internal.StringToBytes("Server: RX\r\n"))

	if len(res.rspBody) > 0 {
		res.body.Write(internal.StringToBytes(fmt.Sprintf("Content-Length: %d", len(res.rspBody)) + "\r\n"))
		res.body.Write(internal.StringToBytes("\r\n"))
		res.body.Write(res.rspBody)
	}
}

func (rsp *responseContext) JSON(status int16, response interface{}) {
	rsp.status = status
	rsp.rspHeaders["Content-Type"] = internal.MIMEJSON
	bits, err := json.Marshal(response)
	if err != nil {
		logger.Log.Error("marshal err: %v", err)
	}
	rsp.rspBody = append(rsp.rspBody, bits...)
}

func (rsp *responseContext) stopTime() time.Time {
	return rsp.time
}

func (rsp *responseContext) SetStopTime(t time.Time) {
	rsp.time = t
}
