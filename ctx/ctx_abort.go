// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type abortContext struct {
	abortStatus int16
	message     string
}

func NewAbortContext(status int16, message string) *abortContext {
	return &abortContext{
		abortStatus: status,
		message:     message,
	}
}
