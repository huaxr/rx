// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package engine

import (
	"net"
	"sync"
	"syscall"
)

func NewPollServer(addr string) *loopServer {
	var err error
	srv := new(loopServer)
	srv.socket, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv.sockFile, err = srv.socket.(*net.TCPListener).File()
	if err != nil {
		panic(err)
	}
	srv.fd = int(srv.sockFile.Fd())
	err = syscall.SetNonblock(srv.fd, true)
	if err != nil {
		panic(err)
	}

	srv.poll = newPoll()
	srv.connections = make(map[int]*conn)

	srv.btsPool = &sync.Pool{
		New: func() interface{} {
			bys := make([]byte, 1<<16)
			return bys
		},
	}
	srv.poll.AddRead(srv.fd)
	return srv
}

func SetKeepAlive(fd, secs int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, 0x8, 1); err != nil {
		return err
	}
	switch err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, 0x101, secs); err {
	case nil, syscall.ENOPROTOOPT: // OS X 10.7 and earlier don't support this option
	default:
		return err
	}
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPALIVE, secs)
}
