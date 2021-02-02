// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package epoll

// +build darwin netbsd freebsd openbsd dragonfly

import (
	"log"
	"syscall"
)

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
	srvFd   int
	changes []syscall.Kevent_t
}

func newPoll() PollIF {
	pl := new(poll)
	p, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}
	pl.srvFd = p
	_, err = syscall.Kevent(pl.srvFd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}
	return pl
}

func (p *poll) ChangeRW(fd int) {
	p.changes = append(p.changes, syscall.Kevent_t{
		Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
	})
}

func (p *poll) ChangeDetach(fd int) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE,
		},
	)
}

func (p *poll) ChangeRead(fd int) {
	p.changes = append(p.changes, syscall.Kevent_t{
		Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE,
	})
}

func (p *poll) AddRW(fd int) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
}

func (p *poll) AddRead(fd int) {
	p.changes = append(p.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
	)
}

// Wait Poll Loop
func (p *poll) Looping(execute func(fd int) error) {
	events := make([]syscall.Kevent_t, 2<<8)
	for {
		n, err := syscall.Kevent(p.srvFd, p.changes, events, nil)
		if err != nil && err != syscall.EINTR {
			log.Println("err while Looping", err)
		}
		p.changes = p.changes[:0]
		for i := 0; i < n; i++ {
			if fd := int(events[i].Ident); fd != 0 {
				if err := execute(fd); err != nil {
					log.Println("err while Looping", err)
				}
			}
		}
	}
}

func (p *poll) Close() error {
	return syscall.Close(p.srvFd)
}

func (p *poll) Trigger(note interface{}) error {
	_, err := syscall.Kevent(p.srvFd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Fflags: syscall.NOTE_TRIGGER,
	}}, nil, nil)
	return err
}
