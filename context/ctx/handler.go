// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type HandlerFunc func(ctx *RequestContext)

var allowed_method = []string{"get", "post"}

func (h *HandlerFunc) Validate(value string) bool {
	for _, i := range allowed_method {
		if value == i {
			return true
		}
	}
	return false
}
