// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package std

import (
	"bufio"
	"bytes"
	"github.com/huaxr/rx/internal"
	"github.com/huaxr/rx/context/ctx"
	"io"
	"log"
	"net"
	"sync"
)

var BtsPool = sync.Pool{
	New: func() interface{} {
		bys := make([]byte, internal.PieceSize)
		return bys
	},
}

type WorkerPool struct {
	size        uint16
	token       chan bool
	connChannel chan net.Conn
	handleErr   chan error
	wg          *sync.WaitGroup
	srv 		*stdServer
}

func NewWorkerPool(size uint16, srv *stdServer) *WorkerPool {
	wp := new(WorkerPool)
	wp.handleErr = make(chan error, 1 << 10)
	go wp.handlerErr()
	wp.size = size
	wp.token = make(chan bool, size)
	wp.wg = new(sync.WaitGroup)
	for i := size; i > 0; i-- {
		wp.token <- true
	}
	wp.connChannel = make(chan net.Conn)
	wp.wg.Add(internal.WorkerSize)
	wp.srv = srv
	for i := internal.WorkerSize; i > 0; i-- {
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
			// add cancel here
			for {
				var buf = BtsPool.Get().([]byte)
				n, err := reader.Read(buf[:])
				if n < internal.PieceSize || err == io.EOF {
					// bug fix
					break
				}
				if err != nil {
					log.Println(err)
					continue
				}
				buffer.Write(buf)
				BtsPool.Put(buf)
			}

			reqContext := ctx.WrapRequest(buffer.Bytes())
			defer ctx.PutContext(reqContext)
			buffer.Reset()

			// set the remoteAddr for logger usage
			reqContext.SetClientAddr(con.RemoteAddr())
			res := reqContext.Execute()
			_, err := con.Write(res.RspToBytes())
			if err != nil {
				log.Println("startWorkers error:", err)
				_ = con.Close()
				continue
			}
			_ = con.Close()
		}
	}
}
