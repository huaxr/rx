// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"net"
	"sync"

	"github.com/huaxr/rx/internal"
)

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
		wrapStd(con)
	}
}
