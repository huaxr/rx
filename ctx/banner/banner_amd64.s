// Copyright 2021 XinRui Hua.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.


// the symbol \ should be replaced by \\ instead.
// reference https://juejin.cn/post/6844903929713524744

// +build amd64 amd64p32 arm arm64 386 ppc64 ppc64le

#include "textflag.h"
DATA  banner<>+0x0000(SB)/48, $" ____   ___ ____  _ /\\/|___  __  __ \n"
DATA  banner<>+0x0030(SB)/48, $"|___ \\ / _ \\___ \\/ |/\\/  _ \\ \\ \\/ / \n"
DATA  banner<>+0x0060(SB)/48, $"  __) | | | |__) | |  | |_) | \\ |/  \n"
DATA  banner<>+0x0090(SB)/48, $"  __) | | | |__) | |  | |_) | \\ |/  \n"
DATA  banner<>+0x00c0(SB)/48, $" / __/| |_| / __/| |  | |_|<  / |\\  \n"
DATA  banner<>+0x00f0(SB)/48, $"|_____|\\___/_____|_|  |_| \\_\\/_/\\_\\ \n"

GLOBL banner<>(SB),NOPTR,$288

TEXT ·PrintBanner(SB), NOSPLIT, $0
	MOVL 	$(0x2000000+4), AX
	MOVQ 	$1, DI
	LEAQ 	banner<>(SB), SI
	MOVL 	$288, DX
	SYSCALL
    RET
