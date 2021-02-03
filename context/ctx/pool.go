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
		return &RequestContext{
			ResponseContext:  ResponseContext{
				rspHeaders: make(map[string]interface{}),
				rspBody: []byte{},
				body: bytes.Buffer{},
			},
			reqHeaders: make(map[string]interface{}),
		}
	},
}

func GetContext() ReqCxtI {
	return reqCtxPool.Get().(ReqCxtI)
}

func PutContext(rc ReqCxtI) {
	rc.Free()
	reqCtxPool.Put(rc)
}
