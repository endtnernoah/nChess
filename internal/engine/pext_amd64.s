// File: pext_amd64.s
#include "textflag.h"

// func pext(x uint64, mask uint64) uint64
TEXT ·Pext(SB), NOSPLIT, $0-24
    MOVQ x+0(FP), BX
    MOVQ mask+8(FP), CX
    PEXTQ CX, BX, AX
    MOVQ AX, ret+16(FP)
    RET
