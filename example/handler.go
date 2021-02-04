// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/huaxr/rx/context/ctx"
)

func handler(ctx ctx.ReqCxtI) {
	name := ctx.GetQuery("name", "XR")
	log.Println(name)
	log.Println("execute handler1")
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "context": name})
}

func handler2(ctx ctx.ReqCxtI) {
	log.Println("execute handler2")
	ctx.Set("user", []string{"a", "b", "c"})
	ctx.Get("user")
}

func handler3(ctx ctx.ReqCxtI) {
	log.Println("execute handler3")
	ctx.Set("a", []string{"a", "b", "c"})
	ctx.Next(handler4)
	return
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "context": "xxxxxx"})
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
	ctx.JSON(200, map[string]interface{}{"time": time.Now(), "context": post.Name})
	log.Println("execute handler5", err, post.Name)
}
