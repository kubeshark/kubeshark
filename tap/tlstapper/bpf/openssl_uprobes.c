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


static __always_inline void ssl_uprobe(struct pt_regs *ctx, void* ssl, void* buffer, int num, struct bpf_map_def* map_fd, size_t *count_ptr) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(map_fd, &id);
	struct ssl_info info = lookup_ssl_info(ctx, &ssl_write_context, id);
	
	info.count_ptr = count_ptr;
	info.buffer = buffer;
	
	long err = bpf_map_update_elem(map_fd, &id, &info, BPF_ANY);
	
	if (err != 0) {
		log_error(ctx, LOG_ERROR_PUTTING_SSL_CONTEXT, id, err, 0l);
	}
}

static __always_inline void ssl_uretprobe(struct pt_regs *ctx, struct bpf_map_def* map_fd, __u32 flags) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(map_fd, &id);
	
	if (infoPtr == NULL) {
		log_error(ctx, LOG_ERROR_GETTING_SSL_CONTEXT, id, 0l, 0l);
		return;
	}
	
	struct ssl_info info;
	long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
	
	// Do not clean map on purpose, sometimes there are two calls to ssl_read in a raw
	//	while the first call actually goes to read from socket, and we get the chance
	//	to find the fd. The other call already have all the information and we don't
	//	have the chance to get the fd.
	//
	// There are two risks keeping the map items
	//	1. It gets full - we solve it by using BPF_MAP_TYPE_LRU_HASH with hard limit
	//	2. We get wrong info of an old call - we solve it by comparing the timestamp 
	//		info before using it
	//
	// bpf_map_delete_elem(map_fd, &id);
	
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SSL_CONTEXT, id, err, ORIGIN_SSL_URETPROBE_CODE);
		return;
	}
	
	if (info.fd == -1) {
		log_error(ctx, LOG_ERROR_MISSING_FILE_DESCRIPTOR, id, 0l, 0l);
		return;
	}

    int count_bytes = get_count_bytes(ctx, &info, id);

	output_ssl_chunk(ctx, &info, count_bytes, id, flags);
}

SEC("uprobe/ssl_write")
void BPF_KPROBE(ssl_write, void* ssl, void* buffer, int num) {
	ssl_uprobe(ctx, ssl, buffer, num, &ssl_write_context, 0);
}

SEC("uretprobe/ssl_write")
void BPF_KPROBE(ssl_ret_write) {
	ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read")
void BPF_KPROBE(ssl_read, void* ssl, void* buffer, int num) {
	ssl_uprobe(ctx, ssl, buffer, num, &ssl_read_context, 0);
}

SEC("uretprobe/ssl_read")
void BPF_KPROBE(ssl_ret_read) {
	ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}

SEC("uprobe/ssl_write_ex")
void BPF_KPROBE(ssl_write_ex, void* ssl, void* buffer, size_t num, size_t *written) {
	ssl_uprobe(ctx, ssl, buffer, num, &ssl_write_context, written);
}

SEC("uretprobe/ssl_write_ex")
void BPF_KPROBE(ssl_ret_write_ex) {
	ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read_ex")
void BPF_KPROBE(ssl_read_ex, void* ssl, void* buffer, size_t num, size_t *readbytes) {
	ssl_uprobe(ctx, ssl, buffer, num, &ssl_read_context, readbytes);
}

SEC("uretprobe/ssl_read_ex")
void BPF_KPROBE(ssl_ret_read_ex) {
	ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}
