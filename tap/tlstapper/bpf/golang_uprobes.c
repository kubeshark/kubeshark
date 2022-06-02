/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/maps.h"


SEC("uprobe/golang_crypto_tls_write")
static __always_inline int golang_crypto_tls_write_uprobe(struct pt_regs *ctx) {
    void* stack_addr = (void*)ctx->rsp;
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    __u32 key_dial;
    // Address at ctx->rsp + 0x20 is common between golang_crypto_tls_write_uprobe and golang_net_http_dialconn_uprobe
    __u32 status = bpf_probe_read(&key_dial, sizeof(key_dial), stack_addr + 0x20);
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_write_uprobe] error reading key_dial: %d", status);
        return 0;
    }

    __u64 key_dial_full = (pid << 32) + key_dial;
    struct socket *s = bpf_map_lookup_elem(&golang_socket_to_write, &key_dial_full);
    if (s == NULL) {
        bpf_printk("[golang_crypto_tls_write_uprobe] error getting socket");
        return 0;
    }

    struct golang_read_write *b = NULL;
    b = bpf_ringbuf_reserve(&golang_read_writes, sizeof(struct golang_read_write), 0);
    if (!b) {
        return 0;
    }

    b->pid = pid;
    b->fd = s->fd;
    // ctx->rsi is common between golang_crypto_tls_write_uprobe and golang_crypto_tls_read_uprobe
    b->conn_addr = ctx->rsi; // go.itab.*net.TCPConn,net.Conn address
    b->is_request = true;
    b->len = ctx->rcx;
    b->cap = ctx->rdi;

    status = bpf_probe_read(&b->data, CHUNK_SIZE, (void*)ctx->rbx);
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_write_uprobe] error reading data: %d", status);
        bpf_ringbuf_discard(b, BPF_RB_FORCE_WAKEUP);
        return 0;
    }

    bpf_ringbuf_submit(b, 0);

    return 0;
}

SEC("uprobe/golang_crypto_tls_read")
static __always_inline int golang_crypto_tls_read_uprobe(struct pt_regs *ctx) {
    void* stack_addr = (void*)ctx->rsp;
    __u64 data_p;
    // Address at ctx->rsp + 0xd8 holds the data
    __u32 status = bpf_probe_read(&data_p, sizeof(data_p), stack_addr + 0xd8);
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_read_uprobe] error reading data pointer: %d", status);
        return 0;
    }

    struct golang_read_write *b = NULL;
    b = bpf_ringbuf_reserve(&golang_read_writes, sizeof(struct golang_read_write), 0);
    if (!b) {
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    b->pid = pid_tgid >> 32;
    // ctx->rsi is common between golang_crypto_tls_write_uprobe and golang_crypto_tls_read_uprobe
    b->conn_addr = ctx->rsi; // go.itab.*net.TCPConn,net.Conn address
    b->is_request = false;
    b->len = ctx->rcx;
    b->cap = ctx->rcx; // no cap info

    status = bpf_probe_read(&b->data, CHUNK_SIZE, (void*)(data_p));
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_read_uprobe] error reading data: %d", status);
        bpf_ringbuf_discard(b, BPF_RB_FORCE_WAKEUP);
        return 0;
    }

    bpf_ringbuf_submit(b, 0);

    return 0;
}

SEC("uprobe/golang_net_socket")
static __always_inline int golang_net_socket_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    // ctx->r14 is common between golang_net_socket_uprobe and golang_net_http_dialconn_uprobe
    __u64 key_socket = (pid << 32) + ctx->r14;
    struct socket *s = bpf_map_lookup_elem(&golang_dial_to_socket, &key_socket);
    if (s == NULL) {
        return 0;
    }

    struct socket b = {
        .pid = s->pid,
        .fd = ctx->rax,
        .key_dial = s->key_dial,
        .conn_addr = 0,
    };

    __u64 key_dial_full = (pid << 32) + s->key_dial;
    __u32 status = bpf_map_update_elem(&golang_socket_to_write, &key_dial_full, &b, BPF_ANY);
    if (status != 0) {
        bpf_printk("[golang_net_socket_uprobe] error updating socket file descriptor: %d", status);
    }

    return 0;
}

SEC("uprobe/golang_net_http_dialconn")
static __always_inline int golang_net_http_dialconn_uprobe(struct pt_regs *ctx) {
    void* stack_addr = (void*)ctx->rsp;
    __u32 key_dial;
    // Address at ctx->rsp + 0x250 is common between golang_crypto_tls_write_uprobe and golang_net_http_dialconn_uprobe
    __u32 status = bpf_probe_read(&key_dial, sizeof(key_dial), stack_addr + 0x250);
    if (status < 0) {
        bpf_printk("[golang_net_http_dialconn_uprobe] error reading key_dial: %d", status);
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct socket b = {
        .pid = pid_tgid >> 32,
        .fd = 0,
        .key_dial = key_dial,
        .conn_addr = 0,
    };

    __u64 pid = b.pid;
    // ctx->r14 is common between golang_net_socket_uprobe and golang_net_http_dialconn_uprobe
    __u64 key_socket = (pid << 32) + ctx->r14;
    status = bpf_map_update_elem(&golang_dial_to_socket, &key_socket, &b, BPF_ANY);
    if (status != 0) {
        bpf_printk("[golang_net_http_dialconn_uprobe] error setting socket: %d", status);
    }

    return 0;
}
