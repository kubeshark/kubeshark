/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"


SEC("uprobe/golang_crypto_tls_write")
static __always_inline int golang_crypto_tls_write_uprobe(struct pt_regs *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 pid = pid_tgid >> 32;
    if (!should_tap(pid)) {
		return 0;
	}

    struct ssl_info *infoPtr = bpf_map_lookup_elem(&ssl_write_context, &pid_tgid);
	struct ssl_info info = {};

	if (infoPtr == NULL) {
		info.fd = -1;
		info.created_at_nano = bpf_ktime_get_ns();
	} else {
		long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);

		if (err != 0) {
			log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, pid_tgid, err, ORIGIN_SSL_UPROBE_CODE);
		}

		if ((bpf_ktime_get_ns() - info.created_at_nano) > SSL_INFO_MAX_TTL_NANO) {
			// If the ssl info is too old, we don't want to use its info because it may be incorrect.
			//
			info.fd = -1;
			info.created_at_nano = bpf_ktime_get_ns();
		}
	}

	info.buffer_len = ctx->rcx;
	info.buffer = (void*)ctx->rbx;

    long err = bpf_map_update_elem(&ssl_write_context, &pid_tgid, &info, BPF_ANY);

	if (err != 0) {
		log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
	}

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, 0);

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

    struct ssl_info *infoPtr = bpf_map_lookup_elem(&ssl_read_context, &pid_tgid);
	struct ssl_info info = {};

	if (infoPtr == NULL) {
		info.fd = -1;
		info.created_at_nano = bpf_ktime_get_ns();
	} else {
		long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);

		if (err != 0) {
			log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, pid_tgid, err, ORIGIN_SSL_UPROBE_CODE);
		}

		if ((bpf_ktime_get_ns() - info.created_at_nano) > SSL_INFO_MAX_TTL_NANO) {
			// If the ssl info is too old, we don't want to use its info because it may be incorrect.
			//
			info.fd = -1;
			info.created_at_nano = bpf_ktime_get_ns();
		}
	}

	info.buffer_len = ctx->rcx;
	info.buffer = (void*)data_p;

    long err = bpf_map_update_elem(&ssl_read_context, &pid_tgid, &info, BPF_ANY);

	if (err != 0) {
		log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, pid_tgid, err, 0l);
	}

    output_ssl_chunk(ctx, &info, info.buffer_len, pid_tgid, FLAGS_IS_READ_BIT);

    return 0;
}
