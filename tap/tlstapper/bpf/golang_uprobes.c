/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include <stdbool.h>

#include "include/headers.h"
#include "include/maps.h"

#define BUFFER_SIZE_READ_WRITE  (1 << 19)  // 512 KiB


struct golang_read_write {
    __u32 pid;
    __u32 fd;
    __u32 conn_addr;
    bool is_request;
    bool is_gzip_chunk;
    __u32 len;
    __u32 cap;
    __u8 data[BUFFER_SIZE_READ_WRITE];
};

const struct golang_read_write *unused __attribute__((unused));


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
    b->is_gzip_chunk = false;
    b->len = ctx->rcx;
    b->cap = ctx->rdi;

    struct socket x = {
        .pid = s->pid,
        .fd = ctx->rax,
        .key_dial = s->key_dial,
        .key_gzip = s->key_gzip,
        .conn_addr = b->conn_addr,
    };

    __u64 key_gzip_full = (pid << 32) + s->key_gzip;
    status = bpf_map_update_elem(&golang_write_to_gzip, &key_gzip_full, &x, BPF_ANY);
    if (status != 0) {
        bpf_printk("[golang_crypto_tls_write_uprobe] error updating gzip conn addr: %d", status);
    }

    status = bpf_probe_read_str(&b->data, sizeof(b->data), (void*)ctx->rbx);
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
    int r14 = ctx->r14;
    // Cancel if it's a gzip read
    if (r14 == 416) {
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
    b->is_gzip_chunk = false;
    b->len = ctx->rcx;
    b->cap = ctx->rcx; // no cap info

    void* stack_addr = (void*)ctx->rsp;
    __u64 data_p;
    // Address at ctx->rsp + 0xd8 holds the data
    __u32 status = bpf_probe_read(&data_p, sizeof(data_p), stack_addr + 0xd8);
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_read_uprobe] error reading data pointer: %d", status);
        bpf_ringbuf_discard(b, BPF_RB_FORCE_WAKEUP);
        return 0;
    }

    status = bpf_probe_read_str(&b->data, sizeof(b->data), (void*)(data_p));
    if (status < 0) {
        bpf_printk("[golang_crypto_tls_read_uprobe] error reading data: %d", status);
        bpf_ringbuf_discard(b, BPF_RB_FORCE_WAKEUP);
        return 0;
    }

    bpf_ringbuf_submit(b, 0);

    return 0;
}

SEC("uprobe/golang_net_http_gzipreader_read")
static __always_inline int golang_net_http_gzipreader_read_uprobe(struct pt_regs *ctx) {
    void* stack_addr = (void*)ctx->rsp;
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    __u32 key_gzip;
    // Address at ctx->rsp + 0x150 is common between golang_net_http_gzipreader_read_uprobe and golang_net_http_dialconn_uprobe
    __u32 status = bpf_probe_read(&key_gzip, sizeof(key_gzip), stack_addr + 0x150);
    if (status < 0) {
        bpf_printk("[golang_net_http_gzipreader_read_uprobe] error reading key_gzip: %d", status);
        return 0;
    }

    __u64 key_gzip_full = (pid << 32) + key_gzip;
    struct socket *s = bpf_map_lookup_elem(&golang_write_to_gzip, &key_gzip_full);
    if (s == NULL) {
        bpf_printk("[golang_net_http_gzipreader_read_uprobe] error getting socket");
        return 0;
    }

    struct golang_read_write *b = NULL;
    b = bpf_ringbuf_reserve(&golang_read_writes, sizeof(struct golang_read_write), 0);
    if (!b) {
        return 0;
    }

    b->pid = pid;
    b->conn_addr = s->conn_addr;
    b->is_request = false;
    b->is_gzip_chunk = true;
    b->len = ctx->rax;
    b->cap = ctx->rax; // no cap info

    __u64 data_p;
    // Address at ctx->rsp + 0x8 holds the data
    status = bpf_probe_read(&data_p, sizeof(data_p), stack_addr + 0x8);
    if (status < 0) {
        bpf_printk("[golang_net_http_gzipreader_read_uprobe] error reading data pointer: %d", status);
        bpf_ringbuf_discard(b, BPF_RB_FORCE_WAKEUP);
        return 0;
    }

    status = bpf_probe_read_str(&b->data, sizeof(b->data), (void*)(data_p));
    if (status < 0) {
        bpf_printk("[golang_net_http_gzipreader_read_uprobe] error reading data: %d", status);
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
        .key_gzip = s->key_gzip,
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

    __u32 key_gzip;
    // Address at ctx->rsp + 0x1d0 is common between golang_net_http_gzipreader_read_uprobe and golang_net_http_dialconn_uprobe
    status = bpf_probe_read(&key_gzip, sizeof(key_gzip), stack_addr + 0x1d0);
    if (status < 0) {
        bpf_printk("[golang_net_http_dialconn_uprobe] error reading key_gzip: %d", status);
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct socket b = {
        .pid = pid_tgid >> 32,
        .fd = 0,
        .key_dial = key_dial,
        .key_gzip = key_gzip,
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
