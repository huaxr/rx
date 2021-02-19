// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package engine

import (
	"net"

	"go.uber.org/atomic"
)

type Signal int

const (
	None Signal = iota
	Detach
	Close
	Shutdown
)

type conn struct {
	// sock is conn socket id
	sock int

	// out, in to handle the input & output
	out []byte
	in  []byte

	// opened atomic value to store the the result
	opened atomic.Bool
	// action for user control signal
	signal Signal

	connInfo net.Addr
}

func (c *conn) isOpen() bool {
	return c.opened.Load()
}
