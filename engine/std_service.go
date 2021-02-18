// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package engine

import (
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/huaxr/rx/ctx"

	"github.com/huaxr/rx/logger"
)

// TcpSocket the raw socket
type stdServer struct {
	//reader io.Reader
	listener net.Listener

	ch []chan net.Conn

	tick  *time.Ticker
	count atomic.Int32

	qps int32
	// runtime.NumGoroutine()
	//num int32

	//stop chan struct{}
	//add  chan struct{}
}

const period = 1
const LB = 1 << 6

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

		ch := make(chan net.Conn, 200)
		for i := 1; i <= LB; i++ {
			t.ch = append(t.ch, ch)
		}
		t.tick = time.NewTicker(period * time.Second)
		//t.stop = make(chan struct{}, 50)
		//t.add = make(chan struct{}, 50)
		go t.calculateQPS()
		t.do()
	})
	return t
}

func (t *stdServer) calculateQPS() {
	for {
		select {
		case <-t.tick.C:
			//last_qps := t.qps
			t.qps = t.count.Load() / period
			//delta := current_qps - last_qps
			//if delta > 0 {
			//	for i := delta; i > 0; i-- {
			//		t.add <- struct{}{}
			//	}
			//} else {
			//	delta = -delta
			//	for i := delta; i > 0; i-- {
			//		t.stop <- struct{}{}
			//	}
			//}
			if t.qps > 0 {
				logger.Log.Info("current qps: %d/s, goroutine count: %d", t.qps, runtime.NumGoroutine())
			}
			t.count.Store(0)
		}
	}
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
		ip := rawConn.RemoteAddr().String()
		ip = strings.ReplaceAll(ip[len(ip)-9:], ".", "")
		ip = strings.ReplaceAll(ip, ":", "")
		lb, _ := strconv.Atoi(ip)
		// multiple channel can enhance the performance
		t.ch[lb%LB] <- rawConn
	}
}

// todo justify
func (t *stdServer) do() {
	for _, channel := range t.ch {
		channel := channel
		go func() {
			for {
				c := <-channel
				ctx.WrapStd(c)
				t.count.Inc()
			}
		}()
	}
}
