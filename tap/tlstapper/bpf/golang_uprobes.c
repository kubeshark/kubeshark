/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.


---

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

`_ex_uprobe` suffixed probes suppose to be `uretprobe`(s) are actually `uprobe`(s)
because of the non-standard ABI of Go. Therefore we probe `ret` mnemonics under the symbol
by automatically finding them through reading the ELF binary and disassembling the symbols.
Disassembly related code located in `golang_offsets.go` file.
Example: We probe an arbitrary point in a function body (offset +559):
https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1296

---

SOURCES:

Tracing Go Functions with eBPF (before 1.17): https://www.grant.pizza/blog/tracing-go-functions-with-ebpf-part-2/
Challenges of BPF Tracing Go: https://blog.0x74696d.com/posts/challenges-of-bpf-tracing-go/
x86 calling conventions: https://en.wikipedia.org/wiki/X86_calling_conventions
Plan 9 from Bell Labs: https://en.wikipedia.org/wiki/Plan_9_from_Bell_Labs
The issue for calling convention change in Go: https://github.com/golang/go/issues/40724
Proposal of Register-based Go calling convention: https://go.googlesource.com/proposal/+/master/design/40724-register-calling.md
Go internal ABI (1.17) specification: https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md
Go internal ABI (current) specification: https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md
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
#include "include/go_types.h"

static __always_inline __u32 get_fd_from_tcp_conn(struct pt_regs *ctx) {
    struct go_interface conn;
    long err = bpf_probe_read(&conn, sizeof(conn), (void*)GO_ABI_INTERNAL_PT_REGS_R1(ctx));
    if (err != 0) {
        return invalid_fd;
    }

    void* net_fd_ptr;
    err = bpf_probe_read(&net_fd_ptr, sizeof(net_fd_ptr), conn.ptr);
    if (err != 0) {
        return invalid_fd;
    }

    __u32 fd;
    err = bpf_probe_read(&fd, sizeof(fd), net_fd_ptr + 0x10);
    if (err != 0) {
        return invalid_fd;
    }

    return fd;
}

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
    info.fd = get_fd_from_tcp_conn(ctx);

    long err = bpf_map_update_elem(&ssl_write_context, &pid_tgid, &info, BPF_ANY);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
    }

    return 0;
}

SEC("uprobe/golang_crypto_tls_write_ex")
static int golang_crypto_tls_write_ex_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return 0;
    }

    struct ssl_info *info_ptr = bpf_map_lookup_elem(&ssl_write_context, &pid_tgid);

    if (info_ptr == NULL) {
        return 0;
    }

    struct ssl_info info;
    long err = bpf_probe_read(&info, sizeof(struct ssl_info), info_ptr);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, pid_tgid, err, ORIGIN_SSL_URETPROBE_CODE);
        return 0;
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

    struct ssl_info info = lookup_ssl_info(ctx, &ssl_read_context, pid_tgid);

    info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R2(ctx);
    info.buffer = (void*)GO_ABI_INTERNAL_PT_REGS_R4(ctx);
    info.fd = get_fd_from_tcp_conn(ctx);

    long err = bpf_map_update_elem(&ssl_read_context, &pid_tgid, &info, BPF_ANY);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
    }

    return 0;
}

SEC("uprobe/golang_crypto_tls_read_ex")
static int golang_crypto_tls_read_ex_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return 0;
    }

    struct ssl_info *info_ptr = bpf_map_lookup_elem(&ssl_read_context, &pid_tgid);

    if (info_ptr == NULL) {
        return 0;
    }

    struct ssl_info info;
    long err = bpf_probe_read(&info, sizeof(struct ssl_info), info_ptr);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, pid_tgid, err, ORIGIN_SSL_URETPROBE_CODE);
        return 0;
    }

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, FLAGS_IS_READ_BIT);

    return 0;
}
