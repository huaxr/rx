// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"net"
	"sync"

	"github.com/huaxr/rx/logger"

	"github.com/huaxr/rx/internal"
)

// TcpSocket the raw socket
type stdServer struct {
	//reader io.Reader
	listener net.Listener
	workers  *WorkerPool
	wg       *sync.WaitGroup
}

func (t *stdServer) GetAddr() string {
	return t.listener.Addr().String()
}

func (t *stdServer) GetNetwork() string {
	return t.listener.Addr().Network()
}

func newServer(network string, addr string) *stdServer {
	t := new(stdServer)
	once := sync.Once{}
	once.Do(func() {
		listen, err := net.Listen(network, addr)
		if err != nil {
			panic(err)
		}
		t.listener = listen
		t.wg = &sync.WaitGroup{}
		t.workers = NewWorkerPool(t)
	})
	return t

}

func (t *stdServer) startServer() {
	logger.Log.Info("start server on: %v", t.GetAddr())
	t.wg.Add(internal.ServerSize)
	for i := internal.ServerSize; i > 0; i-- {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				rawConn, err := t.listener.Accept()
				if err != nil {
					logger.Log.Error("Err listening, %v", err.Error())
					continue
				}
				t.workers.connChannel <- rawConn
			}
		}(t.wg)
	}
	defer func() {
		t.wg.Wait()
		t.workers.wg.Wait()
	}()
}

func NewStdServer(addr string) *stdServer {
	srv := newServer("tcp", addr)
	return srv
}

func (s *stdServer) Run() {
	Print()
	s.startServer()
}

func (s *stdServer) GetWorker() *WorkerPool {
	return s.workers
}
