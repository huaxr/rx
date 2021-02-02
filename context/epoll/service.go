// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package epoll

import (
	"fmt"
	"github.com/huaxr/rx/context/ctx"
	"go.uber.org/atomic"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
)

type loopServer struct {
	socket   net.Listener
	sockFile *os.File
	fd       int
	poll     PollIF
	// connections store the connection socket pair, sync.map
	connections map[int]*conn
	// connections key count
	count   atomic.Int32
	btsPool *sync.Pool
	//wg sync.WaitGroup
	handlers map[string][]ctx.HandlerFunc
	//groupHandlers ctx.HandlerFunc
}

func (srv *loopServer) Run() {
	log.Println("starting the epoll")
	srv.poll.Looping(srv.execute)
}

func (srv *loopServer) Register(method, path string, handlerFunc ...ctx.HandlerFunc) {
	srv.handlers[fmt.Sprintf("%s::", method)+path] = handlerFunc
}

func (s *loopServer) GetHandlers() map[string][]ctx.HandlerFunc {
	return s.handlers
}

func (srv *loopServer) execute(fd int) error {
	c := srv.connections[fd]
	switch {
	case c == nil:
		return srv.loopConnect(fd)
	case !c.isOpen():
		return srv.loopOpened(c)
	case len(c.out) > 0:
		return srv.loopWrite(c)
	case c.signal != None:
		return srv.loopAction(c)
	default:
		return srv.loopRead(c)
	}
}

// loopConnect represent the remote socket connect to this socket
// fd is the unique handle to deal all the raw sock
func (srv *loopServer) loopConnect(fd int) error {
	if srv.fd == fd {
		nfd, sa, err := syscall.Accept(fd)
		if err != nil {
			if err == syscall.EAGAIN {
				return nil
			}
			return err
		}
		if err := syscall.SetNonblock(nfd, true); err != nil {
			return err
		}
		//c := &conn{fd: nfd, sa: sa}
		c := &conn{sock: nfd}

		c.out = nil
		c.in = nil

		c.connInfo = SockaddrToAddr(sa)
		srv.connections[c.sock] = c
		srv.poll.AddRW(c.sock)
		srv.count.Inc()
	}
	return nil
}

func (srv *loopServer) loopOpened(c *conn) error {
	c.opened.Store(true)

	// set keep alive options
	//_ = SetKeepAlive(c.sock, int(10/time.Second))

	if len(c.out) == 0 {
		srv.poll.ChangeRead(c.sock)
	}
	return nil
}

func (srv *loopServer) loopWrite(c *conn) error {
	n, err := syscall.Write(c.sock, c.out)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return srv.closeConn(c)
	}
	c.out = c.out[n:]
	if len(c.out) == 0 {
		srv.poll.ChangeRead(c.sock)
	}
	return nil
}

func (srv *loopServer) closeConn(c *conn) error {
	srv.count.Dec()
	delete(srv.connections, c.sock)
	_ = syscall.Close(c.sock)
	return nil
}

func (srv *loopServer) loopAction(c *conn) error {
	switch c.signal {
	case Close:
		return srv.closeConn(c)
	case Shutdown:
		return nil
	case Detach:
		return srv.loopDetachConn(c)
	default:
		c.signal = None
	}
	if len(c.out) == 0 && c.signal == None {
		srv.poll.ChangeRead(c.sock)
	}

	return nil
}

func (srv *loopServer) loopDetachConn(c *conn) error {
	srv.poll.ChangeDetach(c.sock)
	srv.count.Dec()
	delete(srv.connections, c.sock)
	if err := syscall.SetNonblock(c.sock, false); err != nil {
		return err
	}
	return nil
}

func (srv *loopServer) loopRead(c *conn) error {
	in := srv.btsPool.Get().([]byte)
	defer srv.btsPool.Put(in)

	n, err := syscall.Read(c.sock, in)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return srv.closeConn(c)
	}
	c.in = in[:n]
	// handler write
	if len(c.in) == 1 || string(c.in) == "" {
		return nil
	}

	reqContext := ctx.WrapRequest(c.in)
	reqContext.SetClientAddr(c.connInfo)

	handles := reqContext.GetDefaultHandler()

	if handles == nil {
		var ok bool
		handles, ok = srv.GetHandlers()[reqContext.GetPath(false)]
		if !ok {
			handles = reqContext.GetHandlerByStatus(404)
		}
	}
	res := reqContext.ExecuteSlice(handles...)
	c.out = res.RspToBytes()

	if len(c.out) != 0 || c.signal != None {
		srv.poll.ChangeRW(c.sock)
	}

	defer func() {
		ctx.PutContext(reqContext)
		//c.out = c.out[:0]
	}()

	return nil
}

func (srv *loopServer) Count() int32 {
	return srv.count.Load()
}

func (srv *loopServer) Close() {
	_ = srv.poll.Close()
	_ = srv.socket.Close()
	_ = srv.sockFile.Close()
}

func (srv *loopServer) Use(handlerFunc ...ctx.HandlerFunc) {

}