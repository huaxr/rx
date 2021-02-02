// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import "sync"

var reqCtxPool = sync.Pool{
	New: func() interface{} {
		return &RequestContext{
			reqHeaders: make(map[string]interface{}),
			//params:  make(map[string]interface{}),
			//local : new(sync.Map),
			//response: &ResponseContext{
			//	header:       make(map[string]string),
			//	responseBody: nil,
			//},
		}
	},
}

func GetContext() *RequestContext {
	return reqCtxPool.Get().(*RequestContext)
}

func (rc *RequestContext) PutContext() {
	reqCtxPool.Put(rc)
}
