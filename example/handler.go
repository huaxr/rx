// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/huaxr/rx/ctx"
)

func handler(ctx ctx.ReqCxtI) {
	log.Println("execute handler1")
}

func handler2(ctx ctx.ReqCxtI) {
	log.Println("execute handler2")
}

func handler3(ctx ctx.ReqCxtI) {
	log.Println("execute handler3")
	ctx.Set("a", []string{"a", "b", "c"})
	ctx.Next(handler4)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": "xxxxxx"})
}

func handler4(ctx ctx.ReqCxtI) {
	log.Println("execute handler4")
}

type PostBody struct {
	Name string `json:"name"`
}

func handler5(ctx ctx.ReqCxtI) {
	log.Println(string(ctx.GetBody()))
	var post PostBody
	err := ctx.ParseBody(&post)
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "ctx": post.Name})
	log.Println("execute handler5", err, post.Name)
	ctx.Next(handler)
	//ctx.Abort(200, "sorry")
}
