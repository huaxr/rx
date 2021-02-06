// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import "runtime"

// Stack returns a formatted stack trace of the goroutine that calls it.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
func PrintStack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
