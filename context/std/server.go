// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package std

import (
	"fmt"
	"github.com/huaxr/rx/context/ctx"
	"log"
)

type stdServer struct {
	server   *TcpSocket
	handlers map[string]ctx.HandlerFunc
}

func NewStdServer(addr string) *stdServer {
	core := new(stdServer)
	core.handlers = make(map[string]ctx.HandlerFunc)
	srv := newServer("tcp", addr)
	core.server = srv
	ServerGlobal = core
	return core
}

func (s *stdServer) Run() {
	s.server.startServer()
}

func (s *stdServer) Register(method, path string, handlerFunc ctx.HandlerFunc) {
	if !handlerFunc.Validate(method) {
		log.Println("not support method router")
		return
	}
	s.handlers[fmt.Sprintf("%s::", method)+path] = handlerFunc
}

func (s *stdServer) GetHandlers() map[string]ctx.HandlerFunc {
	return s.handlers
}

func (s *stdServer) GetWorker() *WorkerPool {
	return s.server.workers
}

var ServerGlobal *stdServer
