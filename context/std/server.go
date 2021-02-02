// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package std

import (
	"fmt"
	"github.com/huaxr/rx/context/ctx"
)

func NewStdServer(addr string) *stdServer {
	srv := newServer("tcp", addr)
	srv.handlers = make(map[string][]ctx.HandlerFunc)
	return srv
}

func (s *stdServer) Run() {
	s.startServer()
}

func (s *stdServer) Register(method, path string, handlerFunc ...ctx.HandlerFunc) {
	s.handlers[fmt.Sprintf("%s::", method)+path] = handlerFunc
}

func (s *stdServer) GetHandlers() map[string][]ctx.HandlerFunc {
	return s.handlers
}

func (s *stdServer) GetWorker() *WorkerPool {
	return s.workers
}

func (s *stdServer) Use(handlerFunc ...ctx.HandlerFunc) {

}

