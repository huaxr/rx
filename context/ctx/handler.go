// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type HandlerFunc func(ctx ReqCxtI)

var defaultSTATUS = map[int16]string{
	404 : "Page not found",
	403 : "Bad request",
	500: "Internal server error",
}

var defaultHANDLERS = map[int16]HandlerFunc {
	404 : func(ctx ReqCxtI) {
		ctx.Abort(404, defaultSTATUS[404])
	},
	403 : func(ctx ReqCxtI) {
		ctx.Abort(403, defaultSTATUS[403])
	},
	500 : func(ctx ReqCxtI) {
		ctx.Abort(500, defaultSTATUS[500])
	},
}

func SetDefaultHandler(status int16, handler HandlerFunc) {
	defaultHANDLERS[status] = handler
}

