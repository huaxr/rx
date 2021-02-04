// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"strings"
)

type GroupI interface {
	Register(method, path string, handlerFuncs ...handlerFunc)
	Group(path string, handlerFuncs ...handlerFunc) GroupI
}

type g struct {
	path     string
	handlers []handlerFunc
	pre, next 	*g
}

func Group(path string, handlerFuncs ...handlerFunc) GroupI {
	g := new(g)
	g.handlers = handlerFuncs
	g.pre = nil
	g.next = nil
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.path = path
	return g
}

func (gp *g) Group(path string, handlerFuncs ...handlerFunc) GroupI{
	g := new(g)
	g.handlers = handlerFuncs
	g.pre = gp
	g.next = nil
	return g
}


func (g *g) Register(method, path string, handlerFuncs ...handlerFunc) {
	// when the previous g exist. which means the router has it's
	// group parent may contains handlers to be execute first.
	var preHandlers []handlerFunc
	var prePath string
	for g.pre != nil {
		p := g.pre
		prePath += p.path
		preHandlers = append(preHandlers, p.handlers...)
		p = p.pre
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := prePath + g.path + path
	handlers := []handlerFunc{}
	handlers = append(handlers, g.handlers...)
	handlers = append(handlers, handlerFuncs...)
	Register(method, url, handlers...)
}
