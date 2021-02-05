// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"time"

	"go.uber.org/atomic"
)

type StrategyI interface {
	SetTimeOut(t time.Duration)
	SetTTL(t int32)
}

type strategyContext struct {
	// time to live, stack execute profundity.
	ttl atomic.Int32

	timeout <-chan time.Time

	panic handlerFunc

	openStrategy atomic.Bool
}

// open return the reqCtx field strategyContext
func open() *strategyContext {
	s := new(strategyContext)
	s.openStrategy.Store(false)
	s.timeout = make(<-chan time.Time)
	s.ttl.Store(-1)
	return s
}

func (s *strategyContext) SetTimeOut(t time.Duration) {
	s.openStrategy.Store(true)
	s.timeout = time.After(t)
}

// SetTTL the stack is inc 1 to the t set
func (s *strategyContext) SetTTL(t int32) {
	s.openStrategy.Store(true)
	if t < 0 {
		return
	}
	s.ttl.Store(t + 1)
}

func (s *strategyContext) decTTL() {
	if s.ttl.Load() == 0 {
		return
	}
	s.ttl.Dec()
}

func (s *strategyContext) handleTimeOut(rc *requestContext) {
	rc.setAbort(200, "this router timeout")
}

func (s *strategyContext) handleTTL(rc *requestContext) {
	rc.setAbort(200, "this router ttl out")
}
