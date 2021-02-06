// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"sync"

	"github.com/huaxr/rx/internal"
)

var BtsPool = sync.Pool{
	New: func() interface{} {
		bys := make([]byte, internal.PieceSize)
		return bys
	},
}

type WorkerPool struct {
	size        uint16
	connChannel chan net.Conn
	wg          *sync.WaitGroup
	srv         *stdServer
}

func NewWorkerPool(srv *stdServer) *WorkerPool {
	wp := new(WorkerPool)
	wp.wg = new(sync.WaitGroup)
	wp.connChannel = make(chan net.Conn, 1<<16)
	wp.wg.Add(internal.WorkerSize)
	wp.srv = srv
	for i := internal.WorkerSize; i > 0; i-- {
		go wp.startWorkers()
	}
	return wp
}

func (wp *WorkerPool) startWorkers() {
	//defer worker die
	defer wp.wg.Done()
	for {
		con := <-wp.connChannel
		reader := bufio.NewReader(con)
		var buffer bytes.Buffer
		// add time exceed cancel here
		for {
			var buf = BtsPool.Get().([]byte)
			n, err := reader.Read(buf[:])
			if n < internal.PieceSize || err == io.EOF {
				// eat the last bite
				buffer.Write(buf[:n])
				BtsPool.Put(buf)
				break
			}
			buffer.Write(buf[:n])
			BtsPool.Put(buf)
		}

		reqContext := wrapRequest(buffer.Bytes())
		reqContext.mod = Std
		reqContext.conn = con
		buffer.Reset()
		// set the remoteAddr for logger usage
		reqContext.setClientAddr(con.RemoteAddr())
		reqContext.execute()
		putContext(reqContext)

	}
}
