/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#ifndef __GO_ABI_0__
#define __GO_ABI_0__

/*
Go ABI0 (<=1.16) specification
https://go.dev/doc/asm

Since ABI0 is a stack-based calling convention we only need the stack pointer and
if it's applicable the Goroutine pointer
*/

#include "target_arch.h"

#if defined(bpf_target_x86)

#ifdef __i386__

#define GO_ABI_0_PT_REGS_SP(x) ((x)->esp)

#else

#define GO_ABI_0_PT_REGS_SP(x) ((x)->sp)

#endif

#elif defined(bpf_target_arm)

#define GO_ABI_0_PT_REGS_SP(x) ((x)->uregs[13])
#define GO_ABI_0_PT_REGS_GP(x) ((x)->uregs[10])

#elif defined(bpf_target_arm64)

/* arm64 provides struct user_pt_regs instead of struct pt_regs to userspace */
struct pt_regs;
#define PT_REGS_ARM64 const volatile struct user_pt_regs
#define GO_ABI_0_PT_REGS_SP(x) (((PT_REGS_ARM64 *)(x))->sp)
#define GO_ABI_0_PT_REGS_GP(x) (((PT_REGS_ARM64 *)(x))->regs[18])

#elif defined(bpf_target_powerpc)

#define GO_ABI_0_PT_REGS_SP(x) ((x)->sp)
#define GO_ABI_0_PT_REGS_GP(x) ((x)->gpr[30])

#endif

#endif /* __GO_ABI_0__ */
