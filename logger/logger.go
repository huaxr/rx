// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Logger interface {
	Recovery(format string, val ...interface{})

	Critical(format string, val ...interface{})

	Error(format string, val ...interface{})

	Warning(format string, val ...interface{})

	Info(format string, val ...interface{})
}

var Log Logger

type COLOR int

const (
	BLACK COLOR = iota + 30
	RED
	GREEN
	YELLOW
	BLUE
	PURPLE
)

var colorSet = map[string]COLOR{
	"Recovery": RED,
	"Critical": RED,
	"Error":    PURPLE,
	"Warning":  YELLOW,
	"Info":     GREEN,
}

func InitLogger(l Logger) {
	Log = l
}

func init() {
	dft := new(log)
	Log = dft
}

type log struct {
}

func debugLine() {
	// 1. func name
	_, file, line, ok := runtime.Caller(2)
	if ok {
		path := strings.Split(file, "/rx/")[1] + fmt.Sprintf(":%d", line) + "\n"
		fmt.Fprint(reqWriter, path)
	}
}

func (l *log) do(level string, format string, val ...interface{}) {
	res := fmt.Sprintf("[\u001B[34m%s\u001B[0m] [\u001B[%dm%s\u001B[0m] %s\n",
		time.Now().Format("01/02 15:04:05"), colorSet[level], level, fmt.Sprintf(format, val...))
	_, _ = fmt.Fprint(reqWriter, res)
}

func (l *log) Recovery(format string, val ...interface{}) {
	debugLine()
	l.do("Recovery", format, val...)
}

func (l *log) Critical(format string, val ...interface{}) {
	debugLine()
	l.do("Critical", format, val...)
}

func (l *log) Error(format string, val ...interface{}) {
	debugLine()
	l.do("Error", format, val...)
}

func (l *log) Warning(format string, val ...interface{}) {
	l.do("Warning", format, val...)
}

func (l *log) Info(format string, val ...interface{}) {
	l.do("Info", format, val...)
}
