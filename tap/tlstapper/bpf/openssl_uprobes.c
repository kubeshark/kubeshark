/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#include "include/headers.h"
#include "include/util.h"
#include "include/maps.h"
#include "include/pids.h"

// Heap-like area for eBPF programs - stack size limited to 512 bytes, we must use maps for bigger (chunk) objects.
//
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, int);
	__type(value, struct tlsChunk);
} heap SEC(".maps");

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
		char msg[] = "Error reading bytes count of _ex (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return 0;
	}
	
	return countBytes;
}

static __always_inline void add_address_to_chunk(struct tlsChunk* chunk, __u64 id, __u32 fd) {
	__u32 pid = id >> 32;
	__u64 key = (__u64) pid << 32 | fd;
	
	struct fd_info *fdinfo = bpf_map_lookup_elem(&file_descriptor_to_ipv4, &key);
	
	if (fdinfo == NULL) {
		return;
	}
	
	int err = bpf_probe_read(chunk->address, sizeof(chunk->address), fdinfo->ipv4_addr);
	chunk->flags |= (fdinfo->flags & FLAGS_IS_CLIENT_BIT);
	
	if (err != 0) {
		char msg[] = "Error reading from fd address %ld - %ld";
		bpf_trace_printk(msg, sizeof(msg), id, err);
	}
}

static __always_inline void send_chunk_part(struct pt_regs *ctx, __u8* buffer, __u64 id, 
	struct tlsChunk* chunk, int start, int end) {
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
		char msg[] = "Error reading from ssl buffer %ld - %ld";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	bpf_perf_event_output(ctx, &chunks_buffer, BPF_F_CURRENT_CPU, chunk, sizeof(struct tlsChunk));
}

static __always_inline void send_chunk(struct pt_regs *ctx, __u8* buffer, __u64 id, struct tlsChunk* chunk) {
	// ebpf loops must be bounded at compile time, we can't use (i < chunk->len / CHUNK_SIZE)
	//
	// 	https://lwn.net/Articles/794934/
	// 
	// If we want to compile in kernel older than 5.3, we should add "#pragma unroll" to this loop
	// 
	for (int i = 0; i < MAX_CHUNKS_PER_OPERATION; i++) {
		if (chunk->len <= (CHUNK_SIZE * i)) {
			break;
		}
		
		send_chunk_part(ctx, buffer, id, chunk, CHUNK_SIZE * i, chunk->len);
	}
}

static __always_inline void output_ssl_chunk(struct pt_regs *ctx, struct ssl_info* info, __u64 id, __u32 flags) {
	int countBytes = get_count_bytes(ctx, info, id);
	
	if (countBytes <= 0) {
		return;
	}
	
	if (countBytes > (CHUNK_SIZE * MAX_CHUNKS_PER_OPERATION)) {
		char msg[] = "Buffer too big %d (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), countBytes, id);
		return;
	}
	
	struct tlsChunk* chunk;
	int zero = 0;
	
	// If other thread, running on the same CPU get to this point at the same time like us (context switch)
	//	the data will be corrupted - protection may be added in the future
	//
	chunk = bpf_map_lookup_elem(&heap, &zero);
	
	if (!chunk) {
		char msg[] = "Unable to allocate chunk (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
		return;
	}
	
	chunk->flags = flags;
	chunk->pid = id >> 32;
	chunk->tgid = id;
	chunk->len = countBytes;
	chunk->fd = info->fd;
	
	add_address_to_chunk(chunk, id, chunk->fd);
	send_chunk(ctx, info->buffer, id, chunk);
}

static __always_inline void ssl_uprobe(void* ssl, void* buffer, int num, struct bpf_map_def* map_fd, size_t *count_ptr) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(map_fd, &id);
	struct ssl_info info = {};
	
	if (infoPtr == NULL) {
		info.fd = -1;
		info.created_at_nano = bpf_ktime_get_ns();
	} else {
		long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
		
		if (err != 0) {
			char msg[] = "Error reading old ssl context (id: %ld) (err: %ld)";
			bpf_trace_printk(msg, sizeof(msg), id, err);
		}
		
		if ((bpf_ktime_get_ns() - info.created_at_nano) > SSL_INFO_MAX_TTL_NANO) {
			// If the ssl info is too old, we don't want to use its info because it may be incorrect.
			//
			info.fd = -1;
			info.created_at_nano = bpf_ktime_get_ns();
		}
	}
	
	info.count_ptr = count_ptr;
	info.buffer = buffer;
	
	long err = bpf_map_update_elem(map_fd, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting ssl context (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
	}
}

static __always_inline void ssl_uretprobe(struct pt_regs *ctx, struct bpf_map_def* map_fd, __u32 flags) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(map_fd, &id);
	
	if (infoPtr == NULL) {
		char msg[] = "Error getting ssl context info (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
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
		char msg[] = "Error reading ssl context (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return;
	}
	
	if (info.fd == -1) {
		char msg[] = "File descriptor is missing from ssl info (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
		return;
	}
	
	output_ssl_chunk(ctx, &info, id, flags);
}

SEC("uprobe/ssl_write")
void BPF_KPROBE(ssl_write, void* ssl, void* buffer, int num) {
	ssl_uprobe(ssl, buffer, num, &ssl_write_context, 0);
}

SEC("uretprobe/ssl_write")
void BPF_KPROBE(ssl_ret_write) {
	ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read")
void BPF_KPROBE(ssl_read, void* ssl, void* buffer, int num) {
	ssl_uprobe(ssl, buffer, num, &ssl_read_context, 0);
}

SEC("uretprobe/ssl_read")
void BPF_KPROBE(ssl_ret_read) {
	ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}

SEC("uprobe/ssl_write_ex")
void BPF_KPROBE(ssl_write_ex, void* ssl, void* buffer, size_t num, size_t *written) {
	ssl_uprobe(ssl, buffer, num, &ssl_write_context, written);
}

SEC("uretprobe/ssl_write_ex")
void BPF_KPROBE(ssl_ret_write_ex) {
	ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read_ex")
void BPF_KPROBE(ssl_read_ex, void* ssl, void* buffer, size_t num, size_t *readbytes) {
	ssl_uprobe(ssl, buffer, num, &ssl_read_context, readbytes);
}

SEC("uretprobe/ssl_read_ex")
void BPF_KPROBE(ssl_ret_read_ex) {
	ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}
