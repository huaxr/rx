// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
)

func init() {
	go func() {
		_ = http.ListenAndServe(":9998", nil)
	}()
}

func Init() {
	cpuProfile, _ := os.Create("cpu_profile")
	_ = pprof.StartCPUProfile(cpuProfile)
	memProfile, _ := os.Create("mem_profile")
	_ = pprof.WriteHeapProfile(memProfile)
	defer pprof.StopCPUProfile()
}
