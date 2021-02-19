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

	"github.com/huaxr/rx/ctx"

	"github.com/huaxr/rx/logger"
)

type TYPE string

const (
	HTTP TYPE = "http"
	TCP  TYPE = "tcp"
)

// TcpSocket the raw socket
type stdServer struct {
	//reader io.Reader
	listener net.Listener

	ch []chan net.Conn

	tickQps *time.Ticker
	tickHb  *time.Ticker

	count     int
	qps       int
	goroutine int

	typ TYPE
}

const qpsPeriod = 1
const hbPeriod = 10
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

		ch := make(chan net.Conn, 1000)
		for i := 1; i <= LB; i++ {
			t.ch = append(t.ch, ch)
		}
		t.tickQps = time.NewTicker(qpsPeriod * time.Second)
		t.tickHb = time.NewTicker(hbPeriod * time.Second)
		go t.heartBeat()
		t.do()
	})
	return t
}

func (t *stdServer) heartBeat() {
	for {
		select {
		case <-t.tickQps.C:
			//last_qps := t.qps
			t.qps = t.count / qpsPeriod
			if t.qps > 0 {
				logger.Log.Info("current qps: %d/s", t.qps)
			}
			t.count = 0

		case <-t.tickHb.C:
			g_num := runtime.NumGoroutine()
			t.goroutine = g_num
			logger.Log.Info("current goroutine count: %d", g_num)
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

func (t *stdServer) Type(typ TYPE) {
	t.typ = typ
}

func (t *stdServer) do() {
	for _, channel := range t.ch {
		channel := channel
		go func() {
			for {
				c := <-channel
				ctx.WrapStd(c, string(t.typ))
				t.count++
			}
		}()
	}
}
