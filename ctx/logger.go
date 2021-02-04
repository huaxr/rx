// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"fmt"
	"io"
	"os"
	"time"
)

var DefaultWriter io.Writer = os.Stdout

type LogFormatterParams struct {
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
var defaultLogFormatter = func(param LogFormatterParams) string {
	return fmt.Sprintf("[RX] %v |%3d| %13v | %15s |%-7s %#v\n",
		param.TimeStamp.Format("01/02 15:04:05"),
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method,
		param.Path,
	)
}

func Log(ctx *requestContext) {
	start := ctx.startTime()
	end := ctx.stopTime()
	param := LogFormatterParams{}

	param.TimeStamp = time.Now()
	param.Latency = end.Sub(start)

	param.ClientIP = ctx.GetClientAddr()
	param.Method = ctx.GetMethod()
	param.StatusCode = ctx.getRspStatus()
	param.Path = ctx.getPath(true)

	fmt.Fprint(DefaultWriter, defaultLogFormatter(param))
}
