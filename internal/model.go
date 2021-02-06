// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import "time"

type RequestLogger struct {
	StartTime, StopTime time.Time
	Ip, Method, Path string
	Status int16
}
