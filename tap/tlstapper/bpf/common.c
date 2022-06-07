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


static __always_inline int get_count_bytes(struct pt_regs *ctx, struct ssl_info* info, __u64 id) {
	int returnValue = PT_REGS_RC(ctx);

	if (info->count_ptr == NULL) {
		// ssl_read and ssl_write return the number of bytes written/read
		//
		return returnValue;
	}

	// ssl_read_ex and ssl_write_ex return 1 for success
	//
	if (returnValue != 1) {
		return 0;
	}

	// ssl_read_ex and ssl_write_ex write the number of bytes to an arg named *count
	//
	size_t countBytes;
	long err = bpf_probe_read(&countBytes, sizeof(size_t), (void*) info->count_ptr);

	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_BYTES_COUNT, id, err, 0l);
		return 0;
	}

	return countBytes;
}

static __always_inline int add_address_to_chunk(struct pt_regs *ctx, struct tls_chunk* chunk, __u64 id, __u32 fd) {
	__u32 pid = id >> 32;
	__u64 key = (__u64) pid << 32 | fd;

	struct fd_info *fdinfo = bpf_map_lookup_elem(&file_descriptor_to_ipv4, &key);

	if (fdinfo == NULL) {
		return 0;
	}

	int err = bpf_probe_read(chunk->address, sizeof(chunk->address), fdinfo->ipv4_addr);
	chunk->flags |= (fdinfo->flags & FLAGS_IS_CLIENT_BIT);

	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_FD_ADDRESS, id, err, 0l);
		return 0;
	}

	return 1;
}

static __always_inline void send_chunk_part(struct pt_regs *ctx, __u8* buffer, __u64 id, 
	struct tls_chunk* chunk, int start, int end) {
	size_t recorded = MIN(end - start, sizeof(chunk->data));

	if (recorded <= 0) {
		return;
	}

	chunk->recorded = recorded;
	chunk->start = start;

	// This ugly trick is for the ebpf verifier happiness
	//
	long err = 0;
	if (chunk->recorded == sizeof(chunk->data)) {
		err = bpf_probe_read(chunk->data, sizeof(chunk->data), buffer + start);
	} else {
		recorded &= (sizeof(chunk->data) - 1); // Buffer must be N^2
		err = bpf_probe_read(chunk->data, recorded, buffer + start);
	}

	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_FROM_SSL_BUFFER, id, err, 0l);
		return;
	}

	bpf_perf_event_output(ctx, &chunks_buffer, BPF_F_CURRENT_CPU, chunk, sizeof(struct tls_chunk));
}

static __always_inline void send_chunk(struct pt_regs *ctx, __u8* buffer, __u64 id, struct tls_chunk* chunk) {
	// ebpf loops must be bounded at compile time, we can't use (i < chunk->len / CHUNK_SIZE)
	//
	// 	https://lwn.net/Articles/794934/
	//
	// However we want to run in kernel older than 5.3, hence we use "#pragma unroll" anyway
	//
	#pragma unroll
	for (int i = 0; i < MAX_CHUNKS_PER_OPERATION; i++) {
		if (chunk->len <= (CHUNK_SIZE * i)) {
			break;
		}

		send_chunk_part(ctx, buffer, id, chunk, CHUNK_SIZE * i, chunk->len);
	}
}

static __always_inline void output_ssl_chunk(struct pt_regs *ctx, struct ssl_info* info, int count_bytes, __u64 id, __u32 flags) {
	if (count_bytes <= 0) {
		return;
	}

	if (count_bytes > (CHUNK_SIZE * MAX_CHUNKS_PER_OPERATION)) {
		log_error(ctx, LOG_ERROR_BUFFER_TOO_BIG, id, count_bytes, 0l);
		return;
	}

	struct tls_chunk* chunk;
	int zero = 0;

	// If other thread, running on the same CPU get to this point at the same time like us (context switch)
	//	the data will be corrupted - protection may be added in the future
	//
	chunk = bpf_map_lookup_elem(&heap, &zero);

	if (!chunk) {
		log_error(ctx, LOG_ERROR_ALLOCATING_CHUNK, id, 0l, 0l);
		return;
	}

	chunk->flags = flags;
	chunk->pid = id >> 32;
	chunk->tgid = id;
	chunk->len = count_bytes;
	chunk->fd = info->fd;

	if (!add_address_to_chunk(ctx, chunk, id, chunk->fd)) {
		// Without an address, we drop the chunk because there is not much to do with it in Go
		//
		return;
	}

	send_chunk(ctx, info->buffer, id, chunk);
}
