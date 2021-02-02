// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type HandlerFunc func(ctx ReqCxtI)

var STATUS = map[int16]string{
	404 : "Page not found",
	403 : "Bad request",
	500: "Internal server error",
}

var HANDLERS = map[int16]HandlerFunc {
	404 : func(ctx ReqCxtI) {
		ctx.Abort(404, STATUS[404])
	},
	403 : func(ctx ReqCxtI) {
		ctx.Abort(403, STATUS[403])
	},
	500 : func(ctx ReqCxtI) {
		ctx.Abort(500, STATUS[500])
	},
}
