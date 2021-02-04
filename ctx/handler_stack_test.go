// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"testing"
)

/*goos: darwin
goarch: amd64
pkg: test/test9
Benchmark_Push-4   	10000000	         178 ns/op	      32 B/op	       1 allocs/op
Benchmark_Pop-4    	20000000	        75.5 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	test/test9	9.776s*/

var st *stack

func Init() {
	st = newStack()
}

func Benchmark_Push(b *testing.B) {
	Init()
	for i := 0; i < b.N; i++ { //use b.N for looping
		st.Push(nil)
	}
}

func Benchmark_Pop(b *testing.B) {
	Init()
	b.StopTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		st.Push(nil)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		st.Pop()
	}
}
