// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type AbortI interface {
}

type abortContext struct {
	abortStatus int16
	message     interface{}
}

var defaultSTATUS = map[int16]string{
	404: "Page not found",
	403: "Bad request",
	500: "Internal server error",
}

var defaultHANDLERS = map[int16]handlerFunc{
	404: func(ctx ReqCxtI) {
		ctx.Abort(404, defaultSTATUS[404])
	},
	403: func(ctx ReqCxtI) {
		ctx.Abort(403, defaultSTATUS[403])
	},
	500: func(ctx ReqCxtI) {
		ctx.Abort(500, defaultSTATUS[500])
	},
}

func NewAbortContext(status int16, message interface{}) *abortContext {
	return &abortContext{
		abortStatus: status,
		message:     message,
	}
}
