/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"

struct sys_enter_read_write_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;
	
	__u64 fd;
	__u64* buf;
	__u64 count;
};

struct sys_exit_read_write_ctx {
	__u64 __unused_syscall_header;
	__u32 __unused_syscall_nr;

	__u64 ret;
};


static __always_inline void fd_tracepoints_handle_openssl(struct sys_enter_read_write_ctx *ctx, __u64 id, struct ssl_info *infoPtr, struct bpf_map_def *map_fd, __u64 origin_code) {
	struct ssl_info info;
	long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
	
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, id, err, origin_code);
		return;
	}
	
	info.fd = ctx->fd;
	
	err = bpf_map_update_elem(map_fd, &id, &info, BPF_ANY);
	
	if (err != 0) {
		log_error(ctx, LOG_ERROR_PUTTING_FILE_DESCRIPTOR, id, err, origin_code);
		return;
	}
}

static __always_inline void fd_tracepoints_handle_go(struct sys_enter_read_write_ctx *ctx, __u64 id, struct bpf_map_def *map_fd, __u64 origin_code) {
	__u32 fd = ctx->fd;

	long err = bpf_map_update_elem(map_fd, &id, &fd, BPF_ANY);
	
	if (err != 0) {
		log_error(ctx, LOG_ERROR_PUTTING_FILE_DESCRIPTOR, id, err, origin_code);
		return;
	}
}

SEC("tracepoint/syscalls/sys_enter_read")
void sys_enter_read(struct sys_enter_read_write_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(&openssl_read_context, &id);
	
	if (infoPtr != NULL) {
		fd_tracepoints_handle_openssl(ctx, id, infoPtr, &openssl_read_context, ORIGIN_SYS_ENTER_READ_CODE);
	}

	fd_tracepoints_handle_go(ctx, id, &go_kernel_read_context, ORIGIN_SYS_ENTER_READ_CODE);
}
	
SEC("tracepoint/syscalls/sys_enter_write")
void sys_enter_write(struct sys_enter_read_write_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(&openssl_write_context, &id);
	
	if (infoPtr != NULL) {
		fd_tracepoints_handle_openssl(ctx, id, infoPtr, &openssl_write_context, ORIGIN_SYS_ENTER_WRITE_CODE);
	}

	fd_tracepoints_handle_go(ctx, id, &go_kernel_write_context, ORIGIN_SYS_ENTER_WRITE_CODE);
}

SEC("tracepoint/syscalls/sys_exit_read")
void sys_exit_read(struct sys_exit_read_write_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	// Delete from go map. The value is not used after exiting this syscall.
	// Keep value in openssl map.
	bpf_map_delete_elem(&go_kernel_read_context, &id);
}

SEC("tracepoint/syscalls/sys_exit_write")
void sys_exit_write(struct sys_exit_read_write_ctx *ctx) {
	__u64 id = bpf_get_current_pid_tgid();
	// Delete from go map. The value is not used after exiting this syscall.
	// Keep value in openssl map.
	bpf_map_delete_elem(&go_kernel_write_context, &id);
}
