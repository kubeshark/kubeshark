/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/pids.h"

struct sys_enter_read_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 fd;
	__u64* buf;
	__u64 count;
};

SEC("tracepoint/syscalls/sys_enter_read")
void sys_enter_read(struct sys_enter_read_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(&ssl_read_context, &id);
	
	if (infoPtr == 0) {
		return;
	}
	
	struct ssl_info info;
	long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
	
	if (err != 0) {
		char msg[] = "Error reading read info from read syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	info.fd = ctx->fd;
	
	err = bpf_map_update_elem(&ssl_read_context, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting file descriptor from read syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
}

struct sys_enter_write_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 fd;
	__u64* buf;
	__u64 count;
};

SEC("tracepoint/syscalls/sys_enter_write")
void sys_enter_write(struct sys_enter_write_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(&ssl_write_context, &id);
	
	if (infoPtr == 0) {
		return;
	}
	
	struct ssl_info info;
	long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
	
	if (err != 0) {
		char msg[] = "Error reading write context from write syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	info.fd = ctx->fd;
	
	err = bpf_map_update_elem(&ssl_write_context, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting file descriptor from write syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
}
