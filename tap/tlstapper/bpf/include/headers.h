/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __HEADERS__
#define __HEADERS__

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include "bpf/bpf_tracing.h"

struct user_pt_regs {
	__u64		regs[31];
	__u64		sp;
	__u64		pc;
	__u64		pstate;
};

#endif /* __HEADERS__ */
