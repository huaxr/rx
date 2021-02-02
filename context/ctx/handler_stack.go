// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"sync"
)

type (
	HandlerFuncStack struct {
		top    *node
		length int
		lock   *sync.RWMutex
	}

	node struct {
		value HandlerFunc
		prev  *node
	}
)

// Create a new stack
func NewStack() *HandlerFuncStack {
	return &HandlerFuncStack{nil, 0, &sync.RWMutex{}}
}

// Return the number of items in the stack
func (this *HandlerFuncStack) Len() int {
	return this.length
}

// View the top item on the stack
func (this *HandlerFuncStack) Peek() HandlerFunc {
	if this.length == 0 {
		return nil
	}
	return this.top.value
}

// Pop the top item of the stack and return it
func (this *HandlerFuncStack) Pop() HandlerFunc {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.length == 0 {
		return nil
	}
	n := this.top
	this.top = n.prev
	this.length--
	return n.value
}

// Push a value onto the top of the stack
func (this *HandlerFuncStack) Push(value HandlerFunc) {
	this.lock.Lock()
	defer this.lock.Unlock()
	n := &node{value, this.top}
	this.top = n
	this.length++
}