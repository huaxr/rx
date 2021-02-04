// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"log"
	"net"
	"sync"

	"github.com/huaxr/rx/util"
)

// TcpSocket the raw socket
type stdServer struct {
	//reader io.Reader
	listener net.Listener
	// the channel of socket err, EOF .eg
	acceptErr chan error
	workers   *WorkerPool
	wp        *sync.WaitGroup
	//handlers map[string][]ctx.HandlerFunc
}

func (t *stdServer) GetAddr() string {
	return t.listener.Addr().String()
}

func (t *stdServer) GetNetwork() string {
	return t.listener.Addr().Network()
}

func (t *stdServer) handlerErr() {
	for i := range t.acceptErr {
		log.Println("TcpSocket err:", i)
	}
}

func newServer(network string, addr string) *stdServer {
	t := new(stdServer)
	once := sync.Once{}
	once.Do(func() {
		t.acceptErr = make(chan error, 1024)
		listen, err := net.Listen(network, addr)
		if err != nil {
			panic(err)
		}
		t.listener = listen
		t.workers = NewWorkerPool(util.Concurrent, t)
	})
	go t.handlerErr()
	return t

}

func (t *stdServer) startServer() {
	log.Println("start server on:", t.GetAddr())
	go func() {
		for {
			rawConn, err := t.listener.Accept()
			//_ = rawConn.SetReadDeadline(time.Time{})
			if err != nil {
				t.acceptErr <- err
				continue
			}
			t.workers.PutConnPool(rawConn)
		}
	}()
	t.workers.wg.Wait()
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