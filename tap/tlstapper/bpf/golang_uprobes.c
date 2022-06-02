/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/maps.h"
#include "include/pids.h"
#include "include/log.h"
#include "include/logger_messages.h"


struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, int);
	__type(value, struct golang_read_write);
} golang_heap SEC(".maps");

SEC("uprobe/golang_crypto_tls_write")
static __always_inline int golang_crypto_tls_write_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
		return 0;
	}

    void* stack_addr = (void*)ctx->rsp;
    __u32 key_dial;
    // Address at ctx->rsp + 0x20 is common between golang_crypto_tls_write_uprobe and golang_net_http_dialconn_uprobe
    __u32 status = bpf_probe_read(&key_dial, sizeof(key_dial), stack_addr + 0x20);
    if (status < 0) {
        log_error(ctx, LOG_ERROR_GOLANG_WRITE_READING_KEY_DIAL, pid_tgid, status, 0l);
        return 0;
    }

    __u64 key_dial_full = (pid << 32) + key_dial;
    struct golang_socket *s = bpf_map_lookup_elem(&golang_socket_to_write, &key_dial_full);
    if (s == NULL) {
        log_error(ctx, LOG_ERROR_GOLANG_WRITE_GETTING_SOCKET, pid_tgid, status, 0l);
        return 0;
    }

    struct golang_read_write *b = NULL;
    int zero = 0;

    b = bpf_map_lookup_elem(&golang_heap, &zero);

    if (!b) {
		log_error(ctx, LOG_ERROR_ALLOCATING_CHUNK, pid, 0l, 0l);
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
        log_error(ctx, LOG_ERROR_GOLANG_WRITE_READING_DATA, pid_tgid, status, 0l);
        return 0;
    }

    bpf_perf_event_output(ctx, &golang_read_writes, BPF_F_CURRENT_CPU, b, sizeof(struct golang_read_write));

    return 0;
}

SEC("uprobe/golang_crypto_tls_read")
static __always_inline int golang_crypto_tls_read_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
		return 0;
	}

    void* stack_addr = (void*)ctx->rsp;
    __u64 data_p;
    // Address at ctx->rsp + 0xd8 holds the data
    __u32 status = bpf_probe_read(&data_p, sizeof(data_p), stack_addr + 0xd8);
    if (status < 0) {
        log_error(ctx, LOG_ERROR_GOLANG_READ_READING_DATA_POINTER, pid_tgid, status, 0l);
        return 0;
    }

    struct golang_read_write *b = NULL;
    int zero = 0;

    b = bpf_map_lookup_elem(&golang_heap, &zero);

    if (!b) {
		log_error(ctx, LOG_ERROR_ALLOCATING_CHUNK, pid, 0l, 0l);
		return 0;
	}

    b->pid = pid;
    // ctx->rsi is common between golang_crypto_tls_write_uprobe and golang_crypto_tls_read_uprobe
    b->conn_addr = ctx->rsi; // go.itab.*net.TCPConn,net.Conn address
    b->is_request = false;
    b->len = ctx->rcx;
    b->cap = ctx->rcx; // no cap info

    status = bpf_probe_read(&b->data, CHUNK_SIZE, (void*)(data_p));
    if (status < 0) {
        log_error(ctx, LOG_ERROR_GOLANG_READ_READING_DATA, pid_tgid, status, 0l);
        return 0;
    }

    bpf_perf_event_output(ctx, &golang_read_writes, BPF_F_CURRENT_CPU, b, sizeof(struct golang_read_write));
    return 0;
}

SEC("uprobe/golang_net_socket")
static __always_inline int golang_net_socket_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
		return 0;
	}

    // ctx->r14 is common between golang_net_socket_uprobe and golang_net_http_dialconn_uprobe
    __u64 key_socket = (pid << 32) + ctx->r14;
    struct golang_socket *s = bpf_map_lookup_elem(&golang_dial_to_socket, &key_socket);
    if (s == NULL) {
        log_error(ctx, LOG_ERROR_GOLANG_SOCKET_GETTING_SOCKET, pid_tgid, 0l, 0l);
        return 0;
    }

    struct golang_socket b = {
        .pid = s->pid,
        .fd = ctx->rax,
        .key_dial = s->key_dial,
        .conn_addr = 0,
    };

    __u64 key_dial_full = (pid << 32) + s->key_dial;
    __u32 status = bpf_map_update_elem(&golang_socket_to_write, &key_dial_full, &b, BPF_ANY);
    if (status != 0) {
        log_error(ctx, LOG_ERROR_GOLANG_SOCKET_PUTTING_FILE_DESCRIPTOR, pid_tgid, status, 0l);
    }

    return 0;
}

SEC("uprobe/golang_net_http_dialconn")
static __always_inline int golang_net_http_dialconn_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    if (!should_tap(pid_tgid >> 32)) {
		return 0;
	}

    void* stack_addr = (void*)ctx->rsp;
    __u32 key_dial;
    // Address at ctx->rsp + 0x250 is common between golang_crypto_tls_write_uprobe and golang_net_http_dialconn_uprobe
    __u32 status = bpf_probe_read(&key_dial, sizeof(key_dial), stack_addr + 0x250);
    if (status < 0) {
        log_error(ctx, LOG_ERROR_GOLANG_DIAL_READING_KEY_DIAL, pid_tgid, status, 0l);
        return 0;
    }

    struct golang_socket b = {
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
        log_error(ctx, LOG_ERROR_GOLANG_DIAL_PUTTING_SOCKET, pid_tgid, status, 0l);
    }

    return 0;
}
