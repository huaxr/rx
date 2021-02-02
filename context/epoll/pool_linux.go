// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package epoll

import "syscall"

type PollIF interface {
	ChangeRW(fd int)
	ChangeDetach(fd int)
	ChangeRead(fd int)
	AddRW(fd int)
	AddRead(fd int)
	Looping(execute func(fd int) error)
	Close() error
	Trigger(note interface{}) error
}

type poll struct {
	fd  int // poll fd
	wfd int // wake fd
}

func newPoll() PollIF {
	pl := new(poll)
	p, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	pl.fd = p
	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		syscall.Close(p)
		panic(err)
	}
	pl.wfd = int(r0)
	pl.AddRead(pl.wfd)
	return l
}

func (p *poll) Trigger() error {
	var x uint64 = 1
	_, err := syscall.Write(p.wfd, (*(*[8]byte)(unsafe.Pointer(&x)))[:])
	return err
}

func (p *poll) Wait(iter func(fd int, note interface{}) error) error {
	events := make([]syscall.EpollEvent, 64)
	for {
		n, err := syscall.EpollWait(p.fd, events, 100)
		if err != nil && err != syscall.EINTR {
			return err
		}
		for i := 0; i < n; i++ {
			if fd := int(events[i].Fd); fd != p.wfd {
				if err := iter(fd, nil); err != nil {
					return err
				}
			} else if fd == p.wfd {
				var data [8]byte
				syscall.Read(p.wfd, data[:])
			}
		}
	}
}

func (p *poll) AddRW(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		},
	); err != nil {
		panic(err)
	}
}

func (p *poll) AddRead(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN,
		},
	); err != nil {
		panic(err)
	}
}

func (p *poll) ChangeRead(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN,
		},
	); err != nil {
		panic(err)
	}
}

func (p *poll) ChangeRW(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		},
	); err != nil {
		panic(err)
	}
}

func (p *poll) ChangeDetach(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd,
		&syscall.EpollEvent{Fd: int32(fd),
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		},
	); err != nil {
		panic(err)
	}
}

func (p *poll) Close() error {
	if err := syscall.Close(p.wfd); err != nil {
		return err
	}
	return syscall.Close(p.fd)
}
