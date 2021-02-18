// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/huaxr/rx/internal"
)

func RegisterLogRequest(writer ...io.Writer) {
	reqWriter = io.MultiWriter(writer...)
}

// normal request log
var reqWriter io.Writer = os.Stdout
var enable bool = true

func DisableLog() {
	enable = false
}

type requestFormatter struct {
	// TimeStamp shows log current time
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int16
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP equals Context's ClientIP method.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
}

// defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param requestFormatter) string {
	return fmt.Sprintf("[RX] %v |%3d| %13v | %15s |%-7s %#v\n",
		param.TimeStamp.Format("01/02 15:04:05"),
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method,
		param.Path,
	)
}

func ReqLog(req *internal.RequestLogger) {
	if !enable {
		return
	}
	if reqWriter == nil {
		reqWriter = os.Stdout
	}
	start := req.StartTime
	end := req.StopTime
	param := requestFormatter{}

	param.TimeStamp = time.Now()
	param.Latency = end.Sub(start)

	param.ClientIP = req.Ip
	param.Method = req.Method
	param.StatusCode = req.Status
	param.Path = req.Path

	fmt.Fprint(reqWriter, defaultLogFormatter(param))
}
