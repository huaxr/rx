// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ctx

import "context"

type Context struct {
	Ctx context.Context
	Req ReqCxtI
	Rsp RspCtxI
}
