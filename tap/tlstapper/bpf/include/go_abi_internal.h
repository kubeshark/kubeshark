/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __GOLANG_ABI_INTERNAL__
#define __GOLANG_ABI_INTERNAL__

/*
Go internal ABI specification
https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md
*/

/* Scan the ARCH passed in from ARCH env variable */
#if defined(__TARGET_ARCH_x86)
    #define bpf_target_x86
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_s390)
    #define bpf_target_s390
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_arm)
    #define bpf_target_arm
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_arm64)
    #define bpf_target_arm64
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_mips)
    #define bpf_target_mips
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_powerpc)
    #define bpf_target_powerpc
    #define bpf_target_defined
#elif defined(__TARGET_ARCH_sparc)
    #define bpf_target_sparc
    #define bpf_target_defined
#else
    #undef bpf_target_defined
#endif

/* Fall back to what the compiler says */
#ifndef bpf_target_defined
#if defined(__x86_64__)
    #define bpf_target_x86
#elif defined(__s390__)
    #define bpf_target_s390
#elif defined(__arm__)
    #define bpf_target_arm
#elif defined(__aarch64__)
    #define bpf_target_arm64
#elif defined(__mips__)
    #define bpf_target_mips
#elif defined(__powerpc__)
    #define bpf_target_powerpc
#elif defined(__sparc__)
    #define bpf_target_sparc
#endif
#endif

#if defined(bpf_target_x86)

#ifdef __i386__

/*
https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md#amd64-architecture
https://github.com/golang/go/blob/go1.17.6/src/cmd/compile/internal/ssa/gen/AMD64Ops.go#L100
*/
#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->eax)
#define GO_ABI_INTERNAL_PT_REGS_P2(x) ((x)->ecx)
#define GO_ABI_INTERNAL_PT_REGS_P3(x) ((x)->edx)
#define GO_ABI_INTERNAL_PT_REGS_P4(x) 0
#define GO_ABI_INTERNAL_PT_REGS_P5(x) 0
#define GO_ABI_INTERNAL_PT_REGS_P6(x) 0
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->esp)

#else

#define GO_ABI_INTERNAL_PT_REGS_R1(x) ((x)->rax)
#define GO_ABI_INTERNAL_PT_REGS_R2(x) ((x)->rcx)
#define GO_ABI_INTERNAL_PT_REGS_R3(x) ((x)->rdx)
#define GO_ABI_INTERNAL_PT_REGS_R4(x) ((x)->rbx)
#define GO_ABI_INTERNAL_PT_REGS_R5(x) ((x)->rbp)
#define GO_ABI_INTERNAL_PT_REGS_R6(x) ((x)->rsi)
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->rsp)

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
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->uregs[14])

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
#define GO_ABI_INTERNAL_PT_REGS_SP(x) (((PT_REGS_ARM64 *)(x))->regs[30])

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
#define GO_ABI_INTERNAL_PT_REGS_SP(x) ((x)->sp)

#endif

#endif /* __GOLANG_ABI_INTERNAL__ */
