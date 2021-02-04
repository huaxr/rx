// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"fmt"
	"strings"

	"github.com/huaxr/rx/internal"
)

type HandlerFunc func(ctx ReqCxtI)

type router struct {
	handler []HandlerFunc
	url     string
	method  string
}

var handlerSlice = make(map[int]*router)

func Print() {
	for _, v := range handlerSlice {
		fmt.Printf("\x1b[%dm"+fmt.Sprintf("REGISTER ROUTER: %6s %20s %d", v.method, v.url, len(v.handler))+" \x1b[0m\n", 36)
	}
}

var defaultSTATUS = map[int16]string{
	404: "Page not found",
	403: "Bad request",
	500: "Internal server error",
}

var defaultHANDLERS = map[int16]HandlerFunc{
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

func SetDefaultHandler(status int16, handler HandlerFunc) {
	defaultHANDLERS[status] = handler
}

func Register(method, path string, handlerFuncs ...HandlerFunc) {
	var slice []HandlerFunc
	for l := len(handlerFuncs) - 1; l >= 0; l-- {
		slice = append(slice, handlerFuncs[l])
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	p := internal.CRC(fmt.Sprintf("%s::", method) + path)
	r := new(router)
	r.handler = slice
	r.url = path
	r.method = method
	handlerSlice[p] = r
}
