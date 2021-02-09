// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"testing"
	"time"
)

func Benchmark_Test(b *testing.B) {

}

func TestChan(t *testing.T) {
	ch := time.After(12 * time.Second)
	ch2 := time.After(2 * time.Second)
	for {
		select {
		case <-ch:
			log.Println("exit1")
			return
		case <-ch2:
			log.Println("exi2")
			return
			//default:
			//	log.Println("executing...")
			//	time.Sleep(time.Second * 5)
		}
	}
}
