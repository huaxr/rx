// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"github.com/huaxr/rx/internal"
)

type GroupI interface {
	Register(method, path string, handlerFuncs ...handlerFunc)
	Group(path string, handlerFuncs ...handlerFunc) GroupI
}

type g struct {
	path      string
	handlers  []handlerFunc
	pre, next *g
}

func Group(path string, handlerFuncs ...handlerFunc) GroupI {
	g := new(g)
	g.handlers = handlerFuncs
	// pre equals nil === root
	g.pre = nil
	g.next = nil
	path = internal.CheckPath(path)
	g.path = path
	return g
}

func (gp *g) Group(path string, handlerFuncs ...handlerFunc) GroupI {
	g := new(g)
	g.handlers = handlerFuncs
	g.path = path
	g.pre = gp
	// for search
	gp.next = g
	return g
}

func (g *g) Register(method, path string, handlerFuncs ...handlerFunc) {
	// when the previous g exist. which means the router has it's
	// group parent may contains handlers to be execute first.
	var preHandlers []handlerFunc
	var prePath string

	root := g.pre
	if root != nil {
		// find the root
		for root.pre != nil {
			root = root.pre
		}
		for root.next != nil {
			root.path = internal.CheckPath(root.path)
			prePath += root.path
			preHandlers = append(preHandlers, root.handlers...)
			root = root.next
		}
	}

	path = internal.CheckPath(path)
	g.path = internal.CheckPath(g.path)
	prePath = internal.CheckPath(prePath)

	url := prePath + g.path + path

	handlers := []handlerFunc{}
	handlers = append(handlers, preHandlers...)
	handlers = append(handlers, g.handlers...)
	handlers = append(handlers, handlerFuncs...)
	Register(method, url, handlers...)
}
