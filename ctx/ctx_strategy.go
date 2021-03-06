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
}

type signal struct {
	timeout       bool
	timeoutSignal <-chan time.Time

	//demotion bool
}

type ControlStrategy interface {
	Do() bool
}

// strategy is under developing now, it functions will enhanced later
type StrategyContext struct {
	// time to live, stack execute profundity.
	Ttl int32
	// execute time quota on the stack funcHandlers.
	// if there got a blocking or demanding pending task
	// with the Async true to run it background.

	// set Async identification to inform program to execute
	// the method by asynchronous. TimeOut option is an async true
	Timeout time.Duration

	Async bool

	// degradation the current request with a custom Do option.
	//Demotion ControlStrategy

	// fusing resistor to protect the third part dependency.
	// or other purpose that you made.
	Fusing ControlStrategy

	// for security reason, attack recognition or any other optional
	// functions can be made && check here.
	Security ControlStrategy

	signal
}

// openDefaultStrategy open the default strategy here.
// init the strategyContext fields in order to avoiding the nil pointer.
// consider of the performance, return the reqCtx field strategyContext.
// you can use c.RegisterStrategy(&engine.StrategyContext{Timeout: 1 * time.Second, Ttl: 4})
// to start a StrategyContext on your own.

// default Timeout never expire.
func openDefaultStrategy() *StrategyContext {
	s := new(StrategyContext)
	s.Timeout = 0xff * time.Hour
	s.Ttl = -1
	s.Async = false
	s.timeout = false
	s.timeoutSignal = make(<-chan time.Time)
	return s
}

// wrap the StrategyContext with a default never reached value.
func (s *StrategyContext) wrapDefault() {
	if s.Ttl <= 0 {
		s.Ttl = 0xff
	}
	if s.Timeout <= 0 {
		s.Timeout = 0xff * time.Hour
		s.timeout = false
	} else {
		// when set the strategy, init the timeout chan
		s.timeoutSignal = time.After(s.Timeout)
		s.timeout = true
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

func (s *StrategyContext) decTTL() {
	if s.Ttl == 0 {
		return
	}
	s.Ttl -= 1
}

func (s *StrategyContext) handleTimeOut(rc *RequestContext) {
	rc.setAbort(200, "this router timeout")
}

func (s *StrategyContext) handleTTL(rc *RequestContext) {
	rc.setAbort(200, "this router ttl out")
}

func (s *StrategyContext) handlePanic(rc *RequestContext) {
	rc.setAbort(500, "panic")
}
