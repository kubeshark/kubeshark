/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.



README

Golang does not follow any platform ABI like x86-64 ABI.
Before 1.17, Go followed stack-based Plan9 (Bell Labs) calling convention.
After 1.17, Go switched to an internal register-based calling convention. (Go internal ABI)
The probes in this file supports Go 1.17+

`uretprobe` in Linux kernel uses trampoline pattern to jump to original return
address of the probed function. A Goroutine's stack size is 2Kb while a C thread is 2MB on Linux.
If stack size exceeds 2Kb, Go runtime reallocates the stack. That causes the
return address to become wrong in case of `uretprobe` and probed Go program crashes.
Therefore `uretprobe` CAN'T BE USED for a Go program.

---

SOURCES:

Tracing Go Functions with eBPF (before 1.17): https://www.grant.pizza/blog/tracing-go-functions-with-ebpf-part-2/
Challenges of BPF Tracing Go: https://blog.0x74696d.com/posts/challenges-of-bpf-tracing-go/
x86 calling conventions: https://en.wikipedia.org/wiki/X86_calling_conventions
Plan 9 from Bell Labs: https://en.wikipedia.org/wiki/Plan_9_from_Bell_Labs
The issue for calling convention change in Go: https://github.com/golang/go/issues/40724
Proposal of Register-based Go calling convention: https://go.googlesource.com/proposal/+/master/design/40724-register-calling.md
Go internal ABI (1.17+) specification: https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md
A Quick Guide to Go's Assembler: https://go.googlesource.com/go/+/refs/heads/dev.regabi/doc/asm.html
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"
#include "include/go_abi_internal.h"


SEC("uprobe/golang_crypto_tls_write")
static int golang_crypto_tls_write_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return 0;
    }

    struct ssl_info info = lookup_ssl_info(ctx, &ssl_write_context, pid_tgid);

    info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R2(ctx);
    info.buffer = (void*)GO_ABI_INTERNAL_PT_REGS_R4(ctx);

    long err = bpf_map_update_elem(&ssl_write_context, &pid_tgid, &info, BPF_ANY);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
    }

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, 0);

    return 0;
}

SEC("uprobe/golang_crypto_tls_read")
static int golang_crypto_tls_read_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return 0;
    }

    void* stack_addr = (void*)GO_ABI_INTERNAL_PT_REGS_SP(ctx);
    __u64 data_p;
    // Address at stack pointer + 0xd8 holds the data (*fragile* and probably specific to x86-64)
    __u32 status = bpf_probe_read(&data_p, sizeof(data_p), stack_addr + 0xd8);
    if (status < 0) {
        log_error(ctx, LOG_ERROR_GOLANG_READ_READING_DATA_POINTER, pid_tgid, status, 0l);
        return 0;
    }

    struct ssl_info info = lookup_ssl_info(ctx, &ssl_read_context, pid_tgid);

    info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R2(ctx);
    info.buffer = (void*)data_p;

    long err = bpf_map_update_elem(&ssl_read_context, &pid_tgid, &info, BPF_ANY);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
    }

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, FLAGS_IS_READ_BIT);

    return 0;
}
