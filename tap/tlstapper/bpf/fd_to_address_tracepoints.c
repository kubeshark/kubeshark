/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/pids.h"

struct accept_info {
	__u64* sockaddr;
	__u32* addrlen;
};

BPF_HASH(accept_syscall_context, __u64, struct accept_info);

struct sys_enter_accept4_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 fd;
	__u64* sockaddr;
	__u32* addrlen;
};

SEC("tracepoint/syscalls/sys_enter_accept4")
void sys_enter_accept4(struct sys_enter_accept4_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct accept_info info = {};
	
	info.sockaddr = ctx->sockaddr;
	info.addrlen = ctx->addrlen;
	
	long err = bpf_map_update_elem(&accept_syscall_context, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting accept info (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
}

struct sys_exit_accept4_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 ret;
};

SEC("tracepoint/syscalls/sys_exit_accept4")
void sys_exit_accept4(struct sys_exit_accept4_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	if (ctx->ret < 0) {
		bpf_map_delete_elem(&accept_syscall_context, &id);
		return;
	}
	
	struct accept_info *infoPtr = bpf_map_lookup_elem(&accept_syscall_context, &id);
	
	if (infoPtr == 0) {
		return;
	}
	
	struct accept_info info;
	long err = bpf_probe_read(&info, sizeof(struct accept_info), infoPtr);
	
	bpf_map_delete_elem(&accept_syscall_context, &id);
	
	if (err != 0) {
		char msg[] = "Error reading accept info from accept syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	__u32 addrlen;
	bpf_probe_read(&addrlen, sizeof(__u32), info.addrlen);
	
	if (addrlen != 16) {
		// Currently only ipv4 is supported linux-src/include/linux/inet.h
		return;
	}
	
	struct fd_info fdinfo = {
		.flags = 0
	};
	
	bpf_probe_read(fdinfo.ipv4_addr, sizeof(fdinfo.ipv4_addr), info.sockaddr);
	
	__u32 pid = id >> 32;
	__u32 fd = (__u32) ctx->ret;
	
	__u64 key = (__u64) pid << 32 | fd;
	err = bpf_map_update_elem(&file_descriptor_to_ipv4, &key, &fdinfo, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting fd to address mapping from accept (key: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), key, err);
		return;
	}
}

struct connect_info {
	__u64 fd;
	__u64* sockaddr;
	__u32 addrlen;
};

BPF_HASH(connect_syscall_info, __u64, struct connect_info);

struct sys_enter_connect_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 fd;
	__u64* sockaddr;
	__u32 addrlen;
};

SEC("tracepoint/syscalls/sys_enter_connect")
void sys_enter_connect(struct sys_enter_connect_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct connect_info info = {};
	
	info.sockaddr = ctx->sockaddr;
	info.addrlen = ctx->addrlen;
	info.fd = ctx->fd;
	
	long err = bpf_map_update_elem(&connect_syscall_info, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting connect info (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
}

struct sys_exit_connect_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 ret;
};

SEC("tracepoint/syscalls/sys_exit_connect")
void sys_exit_connect(struct sys_exit_connect_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	// Commented because of async connect which set errno to EINPROGRESS
	// 
	// if (ctx->ret != 0) {
	// 	bpf_map_delete_elem(&accept_syscall_context, &id);
	// 	return;
	// }
	
	struct connect_info *infoPtr = bpf_map_lookup_elem(&connect_syscall_info, &id);
	
	if (infoPtr == 0) {
		return;
	}
	
	struct connect_info info;
	long err = bpf_probe_read(&info, sizeof(struct connect_info), infoPtr);
	
	bpf_map_delete_elem(&connect_syscall_info, &id);
	
	if (err != 0) {
		char msg[] = "Error reading connect info from connect syscall (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	if (info.addrlen != 16) {
		// Currently only ipv4 is supported linux-src/include/linux/inet.h
		return;
	}
	
	struct fd_info fdinfo = {
		.flags = FLAGS_IS_CLIENT_BIT
	};
	
	bpf_probe_read(fdinfo.ipv4_addr, sizeof(fdinfo.ipv4_addr), info.sockaddr);
	
	__u32 pid = id >> 32;
	__u32 fd = (__u32) info.fd;
	
	__u64 key = (__u64) pid << 32 | fd;
	err = bpf_map_update_elem(&file_descriptor_to_ipv4, &key, &fdinfo, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting fd to address mapping from connect (key: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), key, err);
		return;
	}
}
