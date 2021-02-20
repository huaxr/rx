#include "textflag.h"
// https://my.oschina.net/zengsai/blog/123916
TEXT ·Equal(SB),NOSPLIT,$0
        MOVL    len+8(FP), BX
        MOVL    len1+24(FP), CX
        MOVL    $0, AX
        CMPL    BX, CX
        JNE     eqret
        MOVQ    p+0(FP), SI
        MOVQ    q+16(FP), DI
        CLD
        REP; CMPSB
        MOVL    $1, DX
        CMOVLEQ DX, AX
eqret:
        MOVB    AX, ret+32(FP)
        RET

