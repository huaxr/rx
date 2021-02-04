// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

type Server interface {
	Run()
}

func NewServer(mod string, addr string) Server {
	switch mod {
	case "std":
		return NewStdServer(addr)
	case "epoll":
		return NewPollServer(addr)
	default:
		panic("unsupported mod")
	}
	return nil
}