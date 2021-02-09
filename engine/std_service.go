// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package engine

import (
	"net"
	"sync"

	"github.com/huaxr/rx/ctx"

	"github.com/huaxr/rx/logger"
)

// TcpSocket the raw socket
type stdServer struct {
	//reader io.Reader
	listener net.Listener
}

func (t *stdServer) GetAddr() string {
	return t.listener.Addr().String()
}

func NewStdServer(addr string) *stdServer {
	t := new(stdServer)
	once := sync.Once{}
	once.Do(func() {
		listen, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		t.listener = listen
	})
	return t
}

func (t *stdServer) Run() {
	ctx.Print()
	logger.Log.Info("start server on: %v", t.GetAddr())

	for {
		rawConn, err := t.listener.Accept()
		if err != nil {
			logger.Log.Error("Err listening, %v", err.Error())
			continue
		}
		ctx.WrapStd(rawConn)
	}
}
