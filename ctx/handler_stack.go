// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"sync"

	"go.uber.org/atomic"

	"github.com/huaxr/rx/internal"
)

type (
	stack struct {
		top    *node
		length *atomic.Int32
		lock   *sync.RWMutex
	}

	node struct {
		value handlerFunc
		prev  *node
	}
)

// copyStack copy the HandlerFunc from the handlerSlice.
// For each request the has it's own stack to execute
func copyStack(str string) *stack {
	router, ok := handlerSlice[internal.CRC(str)]
	if !ok || len(router.handler) == 0 {
		return nil
	}

	s := newStack()
	for l := len(router.handler) - 1; l >= 0; l-- {
		s.Push(router.handler[l])
	}
	return s
}

// NewStack return the new instance of stack.
// when sync.Poll cache the stack may cause problem here.
func newStack() *stack {
	return &stack{nil, atomic.NewInt32(0), &sync.RWMutex{}}
}

// Len Return the number of items in the stack
func (this *stack) Len() int32 {
	return this.length.Load()
}

// Peek View the top item on the stack
func (this *stack) Peek() handlerFunc {
	if this.length.Load() == 0 {
		return nil
	}
	return this.top.value
}

// Pop the top item of the stack and return it
func (this *stack) Pop() handlerFunc {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.length.Load() == 0 {
		return nil
	}
	n := this.top
	this.top = n.prev
	this.length.Dec()
	return n.value
}

// Push a value onto the top of the stack
func (this *stack) Push(value handlerFunc) {
	this.lock.Lock()
	defer this.lock.Unlock()
	n := &node{value, this.top}
	this.top = n
	this.length.Inc()
}
