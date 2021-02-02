// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"testing"
)

func TestExample(t *testing.T) {
	a := []string{"c"}
	log.Println(a[len(a)-1])
}
