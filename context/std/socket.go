// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package std

import (
	"github.com/huaxr/rx/internal"
	"log"
	"net"
	"sync"
)

// TcpSocket the raw socket
type TcpSocket struct {
	//reader io.Reader
	listener net.Listener

	// the channel of socket err, EOF .eg
	acceptErr chan error
	workers   *WorkerPool

	wp *sync.WaitGroup
}

func (t *TcpSocket) GetAddr() string {
	return t.listener.Addr().String()
}

func (t *TcpSocket) GetNetwork() string {
	return t.listener.Addr().Network()
}

func (t *TcpSocket) handlerErr() {
	for i := range t.acceptErr {
		log.Println("TcpSocket err:", i)
	}
}

func newServer(network string, addr string) *TcpSocket {
	t := new(TcpSocket)
	once := sync.Once{}
	once.Do(func() {
		t.acceptErr = make(chan error, 1024)
		listen, err := net.Listen(network, addr)
		if err != nil {
			panic(err)
		}
		t.listener = listen
		t.workers = NewWorkerPool(internal.Concurrent)
	})
	go t.handlerErr()
	return t

}

func (t *TcpSocket) startServer() {
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
