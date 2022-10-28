/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#ifndef __TARGET_ARCH__
#define __TARGET_ARCH__

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

#endif /* __TARGET_ARCH__ */
