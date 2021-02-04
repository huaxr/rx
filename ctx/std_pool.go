// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"sync"

	"github.com/huaxr/rx/util"
)

var BtsPool = sync.Pool{
	New: func() interface{} {
		bys := make([]byte, util.PieceSize)
		return bys
	},
}

type WorkerPool struct {
	size        uint16
	token       chan bool
	connChannel chan net.Conn
	handleErr   chan error
	wg          *sync.WaitGroup
	srv         *stdServer
}

func NewWorkerPool(size uint16, srv *stdServer) *WorkerPool {
	wp := new(WorkerPool)
	wp.handleErr = make(chan error, 1<<10)
	go wp.handlerErr()
	wp.size = size
	wp.token = make(chan bool, size)
	wp.wg = new(sync.WaitGroup)
	for i := size; i > 0; i-- {
		wp.token <- true
	}
	wp.connChannel = make(chan net.Conn)
	wp.wg.Add(util.WorkerSize)
	wp.srv = srv
	for i := util.WorkerSize; i > 0; i-- {
		go wp.startWorkers()
	}
	return wp
}

func (t *WorkerPool) handlerErr() {
	for err := range t.handleErr {
		if err != nil {
			log.Println("WorkerPool err:", err)
		}
	}
}

func (wp *WorkerPool) PutConnPool(conn net.Conn) {
	<-wp.token
	wp.connChannel <- conn
	wp.token <- true
}

func (wp *WorkerPool) startWorkers() {
	//defer worker die
	defer wp.wg.Done()
	for {
		select {
		case con := <-wp.connChannel:
			reader := bufio.NewReader(con)
			var buffer bytes.Buffer
			// add time exceed cancel here
			for {
				var buf = BtsPool.Get().([]byte)
				n, err := reader.Read(buf[:])
				if n < util.PieceSize || err == io.EOF {
					// eat the last bite
					buffer.Write(buf[:n])
					BtsPool.Put(buf)
					break
				}
				buffer.Write(buf[:n])
				BtsPool.Put(buf)
			}

			reqContext := wrapRequest(buffer.Bytes())
			defer PutContext(reqContext)
			buffer.Reset()

			// set the remoteAddr for logger usage
			reqContext.setClientAddr(con.RemoteAddr())
			res := reqContext.execute()
			_, err := con.Write(res.rspToBytes())
			if err != nil {
				log.Println("startWorkers error:", err)
				_ = con.Close()
				continue
			}
			_ = con.Close()
		}
	}
}
