// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bytes"
	"sync"
)

var reqCtxPool = sync.Pool{
	New: func() interface{} {
		return &requestContext{
			responseContext: &responseContext{
				rspHeaders: make(map[string]interface{}),
				rspBody:    []byte{},
				body:       bytes.Buffer{},
			},
			reqHeaders: make(map[string]interface{}),
		}
	},
}

func GetContext() *requestContext {
	return reqCtxPool.Get().(*requestContext)
}

func PutContext(rc *requestContext) {
	rc.free()
	reqCtxPool.Put(rc)
}
