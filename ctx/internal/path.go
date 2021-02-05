// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import "strings"

func AddPrefix(path string, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		path = prefix + path
	}
	return path
}

func TrimSuffix(path string, suffix string) string {
	if strings.HasSuffix(path, suffix) {
		return path[:len(path)-1]
	}
	return path
}

func CheckPath(path string) string {
	path = AddPrefix(path, "/")
	path = TrimSuffix(path, "/")
	return path
}
