// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
)

type Logger interface {

	Critical(message ...string)

	Error(message ...string)

	Warning(message ...string)

	Info(message ...string)
}

var Log Logger


func InitLogger(l Logger) {
	Log = l
}

func init() {
	dft := new(log)
	Log = dft
}

type log struct {

}

func (l *log) do(level string, message ...string) {
	if len(message) == 1 {
		fmt.Fprint(reqWriter, fmt.Sprintf("[RX %s]: %s", level, message))
	} else {
		mes := message[0]
		fmt.Fprint(reqWriter, fmt.Sprintf("[RX %s]: %s", level, fmt.Sprintf(mes, message[0:])))
	}
}

func (l *log) Critical(message ...string) {
	l.do("Critical", message...)
}

func (l *log) Error(message ...string) {
	l.do("Error", message...)
}

func (l *log) Warning(message ...string) {
	l.do("Warning", message...)
}

func (l *log) Info(message ...string) {
	l.do("Info", message...)
}