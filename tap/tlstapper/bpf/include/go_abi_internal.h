/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#ifndef __GO_ABI_INTERNAL__
#define __GO_ABI_INTERNAL__

/*
Go internal ABI (1.17/current) specification
https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md
*/

#include "target_arch.h"

#if defined(bpf_target_x86)

#ifdef __i386__

/*
https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md#amd64-architecture
https://github.com/golang/go/blob/go1.17.6/src/cmd/compile/internal/ssa/gen/AMD64Ops.go#L100
*/
#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->eax)
#define GO_ABI_INTERNAL_PT_REGS_R2(x) ((x)->ecx)
#define GO_ABI_INTERNAL_PT_REGS_R3(x) ((x)->edx)
#define GO_ABI_INTERNAL_PT_REGS_R4(x) 0
#define GO_ABI_INTERNAL_PT_REGS_R5(x) 0
#define GO_ABI_INTERNAL_PT_REGS_R6(x) 0
#define GO_ABI_INTERNAL_PT_REGS_R7(x) 0
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->esp)
#define GO_ABI_INTERNAL_PT_REGS_FP(x) ((x)->ebp)
#define GO_ABI_INTERNAL_PT_REGS_GP(x) ((x)->e14)

#else

#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->ax)
#define GO_ABI_INTERNAL_PT_REGS_R2(x) ((x)->cx)
#define GO_ABI_INTERNAL_PT_REGS_R3(x) ((x)->dx)
#define GO_ABI_INTERNAL_PT_REGS_R4(x) ((x)->bx)
#define GO_ABI_INTERNAL_PT_REGS_R5(x) ((x)->bp)
#define GO_ABI_INTERNAL_PT_REGS_R6(x) ((x)->si)
#define GO_ABI_INTERNAL_PT_REGS_R7(x) ((x)->di)
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->sp)
#define GO_ABI_INTERNAL_PT_REGS_FP(x) ((x)->bp)
#define GO_ABI_INTERNAL_PT_REGS_GP(x) ((x)->r14)

#endif

#elif defined(bpf_target_arm)

/*
https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md#arm64-architecture
https://github.com/golang/go/blob/go1.17.6/src/cmd/compile/internal/ssa/gen/ARM64Ops.go#L129-L131
*/
#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->uregs[0])
#define GO_ABI_INTERNAL_PT_REGS_R2(x) ((x)->uregs[1])
#define GO_ABI_INTERNAL_PT_REGS_R3(x) ((x)->uregs[2])
#define GO_ABI_INTERNAL_PT_REGS_R4(x) ((x)->uregs[3])
#define GO_ABI_INTERNAL_PT_REGS_R5(x) ((x)->uregs[4])
#define GO_ABI_INTERNAL_PT_REGS_R6(x) ((x)->uregs[5])
#define GO_ABI_INTERNAL_PT_REGS_R7(x) ((x)->uregs[6])
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->uregs[13])
#define GO_ABI_INTERNAL_PT_REGS_FP(x) ((x)->uregs[29])
#define GO_ABI_INTERNAL_PT_REGS_GP(x) ((x)->uregs[28])

#elif defined(bpf_target_arm64)

/* arm64 provides struct user_pt_regs instead of struct pt_regs to userspace */
struct pt_regs;
#define PT_REGS_ARM64 const volatile struct user_pt_regs
#define GO_ABI_INTERNAL_PT_REGS_R1(x) (((PT_REGS_ARM64 *)(x))->regs[0])
#define GO_ABI_INTERNAL_PT_REGS_R2(x) (((PT_REGS_ARM64 *)(x))->regs[1])
#define GO_ABI_INTERNAL_PT_REGS_R3(x) (((PT_REGS_ARM64 *)(x))->regs[2])
#define GO_ABI_INTERNAL_PT_REGS_R4(x) (((PT_REGS_ARM64 *)(x))->regs[3])
#define GO_ABI_INTERNAL_PT_REGS_R5(x) (((PT_REGS_ARM64 *)(x))->regs[4])
#define GO_ABI_INTERNAL_PT_REGS_R6(x) (((PT_REGS_ARM64 *)(x))->regs[5])
#define GO_ABI_INTERNAL_PT_REGS_R7(x) (((PT_REGS_ARM64 *)(x))->regs[6])
#define GO_ABI_INTERNAL_PT_REGS_SP(x) (((PT_REGS_ARM64 *)(x))->sp)
#define GO_ABI_INTERNAL_PT_REGS_FP(x) (((PT_REGS_ARM64 *)(x))->regs[29])
#define GO_ABI_INTERNAL_PT_REGS_GP(x) (((PT_REGS_ARM64 *)(x))->regs[28])

#elif defined(bpf_target_powerpc)

/*
https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md#ppc64-architecture
https://github.com/golang/go/blob/go1.17.6/src/cmd/compile/internal/ssa/gen/PPC64Ops.go#L125-L127
*/
#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->gpr[3])
#define GO_ABI_INTERNAL_PT_REGS_R2(x) ((x)->gpr[4])
#define GO_ABI_INTERNAL_PT_REGS_R3(x) ((x)->gpr[5])
#define GO_ABI_INTERNAL_PT_REGS_R4(x) ((x)->gpr[6])
#define GO_ABI_INTERNAL_PT_REGS_R5(x) ((x)->gpr[7])
#define GO_ABI_INTERNAL_PT_REGS_R6(x) ((x)->gpr[8])
#define GO_ABI_INTERNAL_PT_REGS_R7(x) ((x)->gpr[9])
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->sp)
#define GO_ABI_INTERNAL_PT_REGS_FP(x) ((x)->gpr[12])
#define GO_ABI_INTERNAL_PT_REGS_GP(x) ((x)->gpr[30])

#endif

#endif /* __GO_ABI_INTERNAL__ */
