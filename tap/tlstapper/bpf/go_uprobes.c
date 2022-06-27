/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.


---

README

Go does not follow any platform ABI like x86-64 System V ABI.
Before 1.17, Go followed stack-based Plan9 (Bell Labs) calling convention. (ABI0)
After 1.17, Go switched to an internal register-based calling convention. (ABIInternal)
For now, the probes in this file supports only ABIInternal (Go 1.17+)

`uretprobe` in Linux kernel uses trampoline pattern to jump to original return
address of the probed function. A Goroutine's stack size is 2Kb while a C thread is 2MB on Linux.
If stack size exceeds 2Kb, Go runtime relocates the stack. That causes the
return address to become incorrect in case of `uretprobe` and probed Go program crashes.
Therefore `uretprobe` CAN'T BE USED for a Go program.

`_ex_uprobe` suffixed probes suppose to be `uretprobe`(s) are actually `uprobe`(s)
because of the non-standard ABI of Go. Therefore we probe all `ret` mnemonics under the symbol
by automatically finding them through reading the ELF binary and disassembling the symbols.
Disassembly related code located in `go_offsets.go` file and it uses Capstone Engine.
Solution based on: https://github.com/iovisor/bcc/issues/1320#issuecomment-407927542
*Example* We probe an arbitrary point in a function body (offset +559):
https://github.com/golang/go/blob/go1.17.6/src/crypto/tls/conn.go#L1299

We get the file descriptor using the common $rax register that holds the address
of `go.itab.*net.TCPConn,net.Conn` and through a series of dereferencing
using `bpf_probe_read` calls in `go_crypto_tls_get_fd_from_tcp_conn` function.

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
Dissecting Go Binaries: https://www.grant.pizza/blog/dissecting-go-binaries/
Capstone Engine: https://www.capstone-engine.org/
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

static __always_inline __u32 go_crypto_tls_get_fd_from_tcp_conn(struct pt_regs *ctx) {
    struct go_interface conn;
    long err;
    __u64 addr;
#if defined(bpf_target_arm64)
    err = bpf_probe_read(&addr, sizeof(addr), (void*)GO_ABI_INTERNAL_PT_REGS_SP(ctx)+0x8);
    if (err != 0) {
        return invalid_fd;
    }
#else
    addr = GO_ABI_INTERNAL_PT_REGS_R1(ctx);
#endif

    err = bpf_probe_read(&conn, sizeof(conn), (void*)addr);
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

static __always_inline void go_crypto_tls_uprobe(struct pt_regs *ctx, struct bpf_map_def* go_context) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return;
    }

    struct ssl_info info = new_ssl_info();
    long err;

#if defined(bpf_target_arm64)
    err = bpf_probe_read(&info.buffer_len, sizeof(__u32), (void*)GO_ABI_INTERNAL_PT_REGS_SP(ctx)+0x18);
    if (err != 0) {
        log_error(ctx, LOG_ERROR_READING_BYTES_COUNT, pid_tgid, err, ORIGIN_SSL_UPROBE_CODE);
        return;
    }
#else
    info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R2(ctx);
#endif
    info.buffer = (void*)GO_ABI_INTERNAL_PT_REGS_R4(ctx);
    info.fd = go_crypto_tls_get_fd_from_tcp_conn(ctx);

    // GO_ABI_INTERNAL_PT_REGS_GP is Goroutine address
    __u64 pid_fp = pid << 32 | GO_ABI_INTERNAL_PT_REGS_GP(ctx);
    err = bpf_map_update_elem(go_context, &pid_fp, &info, BPF_ANY);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
    }

    return;
}

static __always_inline void go_crypto_tls_ex_uprobe(struct pt_regs *ctx, struct bpf_map_def* go_context, __u32 flags) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
        return;
    }

    // GO_ABI_INTERNAL_PT_REGS_GP is Goroutine address
    __u64 pid_fp = pid << 32 | GO_ABI_INTERNAL_PT_REGS_GP(ctx);
    struct ssl_info *info_ptr = bpf_map_lookup_elem(go_context, &pid_fp);

    if (info_ptr == NULL) {
        return;
    }
    bpf_map_delete_elem(go_context, &pid_fp);

    struct ssl_info info;
    long err = bpf_probe_read(&info, sizeof(struct ssl_info), info_ptr);

    if (err != 0) {
        log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, pid_tgid, err, ORIGIN_SSL_URETPROBE_CODE);
        return;
    }

    // In case of read, the length is determined on return
    if (flags == FLAGS_IS_READ_BIT) {
#if defined(bpf_target_arm64)
        // On ARM64 we look at a general-purpose register as an indicator of error return
        if (GO_ABI_INTERNAL_PT_REGS_R6(ctx) == 0x10) {
            return;
        }
        info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R7(ctx); // n in return n, nil
#else
        info.buffer_len = GO_ABI_INTERNAL_PT_REGS_R1(ctx); // n in return n, nil
#endif
        // This check achieves ignoring 0 length reads (the reads result with an error)
        if (info.buffer_len <= 0) {
            return;
        }
    }

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, flags);

    return;
}

SEC("uprobe/go_crypto_tls_write")
void BPF_KPROBE(go_crypto_tls_write) {
    go_crypto_tls_uprobe(ctx, &go_write_context);
}

SEC("uprobe/go_crypto_tls_write_ex")
void BPF_KPROBE(go_crypto_tls_write_ex) {
    go_crypto_tls_ex_uprobe(ctx, &go_write_context, 0);
}

SEC("uprobe/go_crypto_tls_read")
void BPF_KPROBE(go_crypto_tls_read) {
    go_crypto_tls_uprobe(ctx, &go_read_context);
}

SEC("uprobe/go_crypto_tls_read_ex")
void BPF_KPROBE(go_crypto_tls_read_ex) {
    go_crypto_tls_ex_uprobe(ctx, &go_read_context, FLAGS_IS_READ_BIT);
}
