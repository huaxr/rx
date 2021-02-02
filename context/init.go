// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package context

import (
	"github.com/huaxr/rx/context/ctx"
	"github.com/huaxr/rx/context/epoll"
	"github.com/huaxr/rx/context/std"
)

type Server interface {
	Run()
	// Register handlers to the srv
	Register(method, path string, handlerFunc ctx.HandlerFunc)
	GetHandlers() map[string]ctx.HandlerFunc
}

func NewServer(mod string, addr string) Server {
	switch mod {
	case "std":
		return std.NewStdServer(addr)
	case "epoll":
		return epoll.NewPollServer(addr)
	default:
		panic("unsupported mod")
	}
	return nil
}
