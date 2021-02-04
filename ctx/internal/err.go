// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

type MyError struct {
	description string
	code        int
}

func (e *MyError) Error() string {
	return e.description
}
func (e *MyError) Code() int {
	return e.code
}

var NoRecord = &MyError{code: 1, description: "record not exists"}
var TokenInvalid = &MyError{code: 2, description: "token invalid"}

const IOCGetterErrorCode = 4
