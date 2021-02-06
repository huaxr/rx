// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import (
	"time"
)

type StrategyI interface {
	SetTimeOut(t time.Duration)
	SetTTL(t int32)
	SetAsync(t bool)
}

type PanicI interface {
	Do()
}

type signal struct {
	// the signal definition
	asyncSignal chan bool
	timeoutSignal <-chan time.Time
}

// strategy is under developing now, it functions will enhanced later
type StrategyContext struct {
	// time to live, stack execute profundity.
	Ttl int32
	// execute time quota on the stack funcHandlers.
	// if there got a blocking or demanding pending task
	// with the Async true to run it background.
	Timeout time.Duration
	// set Async identification to inform program to execute
	// the method by asynchronous
	Async bool

	// panic happens in execute function
	Panic PanicI

	signal
}

// openDefaultStrategy open the default strategy here.
// init the strategyContext fields in order to avoiding the nil pointer.
// consider of the performance, return the reqCtx field strategyContext.
// you can use c.RegisterStrategy(&ctx.StrategyContext{Timeout: 1 * time.Second, Ttl: 4})
// to start a StrategyContext on your own.

// default Timeout never expire.
func openDefaultStrategy() *StrategyContext {
	s := new(StrategyContext)
	s.Timeout = 0xff * time.Hour
	s.Ttl = -1
	s.Async = false
	s.asyncSignal = make(chan bool, 1)
	s.timeoutSignal = make(<-chan time.Time, 1)
	return s
}

// wrap the StrategyContext with a default never reached value.
func (s *StrategyContext) wrapDefault() {
	if s.Ttl <= 0 {
		s.Ttl = 0xff
	}
	if s.Timeout <= 0 {
		s.Timeout = 0xff * time.Hour
	}
}

func (s *StrategyContext) SetTimeOut(t time.Duration) {
	s.timeoutSignal = time.After(t)
}

// SetTTL the stack is inc 1 to the t set
func (s *StrategyContext) SetTTL(t int32) {
	if t < 0 {
		return
	}
	s.Ttl = t + 1
}

func (s *StrategyContext) SetAsync(t bool) {
	s.Async = t
}

func (s *StrategyContext) decTTL() {
	if s.Ttl == 0 {
		return
	}
	s.Ttl -= 1
}

func (s *StrategyContext) handleTimeOut(rc *requestContext) {
	rc.setAbort(200, "this router timeout")
}

func (s *StrategyContext) handleTTL(rc *requestContext) {
	rc.setAbort(200, "this router ttl out")
}

func (s *StrategyContext) handlePanic(rc *requestContext) {
	rc.setAbort(500, "panic")
}
