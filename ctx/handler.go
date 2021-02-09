// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"fmt"
	"strings"

	"github.com/huaxr/rx/internal"
)

//type IsAsync bool
//
//
//type handlerCore struct {
//	async IsAsync
//	handlerFunc func(engine ReqCxtI)
//	handlerAsyncFunc func(engine ReqCxtI) (done chan bool)
//}

type handlerFunc func(ctx ReqCxtI)

type router struct {
	handler []handlerFunc
	url     string
	method  string
}

var handlerSlice = make(map[int]*router)

func Print() {
	for _, v := range handlerSlice {
		fmt.Printf("\x1b[%dm"+fmt.Sprintf("|EGISTER ROUTER:| %6s |%20s |%d|", v.method, v.url, len(v.handler))+" \x1b[0m\n", 36)
	}
}

func Register(method, path string, handlerFuncs ...handlerFunc) {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	p := internal.CRC(fmt.Sprintf("%s::", method) + path)
	r := new(router)
	r.handler = handlerFuncs
	r.url = path
	r.method = method
	handlerSlice[p] = r
}
