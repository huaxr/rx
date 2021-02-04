// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"strings"
)

type g struct {
	path     string
	handlers []handlerFunc
}

func Group(path string, handlerFuncs ...handlerFunc) *g {
	g := new(g)
	g.handlers = handlerFuncs

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	g.path = path
	return g
}

func (g *g) Register(method, path string, handlerFuncs ...handlerFunc) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := g.path + path
	handlers := []handlerFunc{}
	handlers = append(handlers, g.handlers...)
	handlers = append(handlers, handlerFuncs...)
	Register(method, url, handlers...)
}