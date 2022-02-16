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

static __always_inline int ssl_uprobe(void* ssl, void* buffer, int num, struct bpf_map_def* map_fd, size_t *count_ptr) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return 0;
	}
	
	struct ssl_info info = {};
	
	info.fd = -1;
	info.count_ptr = count_ptr;
	info.buffer = buffer;
	
	long err = bpf_map_update_elem(map_fd, &id, &info, BPF_ANY);
	
	if (err != 0) {
		char msg[] = "Error putting ssl context (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return 0;
	}
	
	return 0;
}

static __always_inline int ssl_uretprobe(struct pt_regs *ctx, struct bpf_map_def* map_fd, __u32 flags) {
	__u64 id = bpf_get_current_pid_tgid();
	
	if (!should_tap(id >> 32)) {
		return 0;
	}
	
	struct ssl_info *infoPtr = bpf_map_lookup_elem(map_fd, &id);
	
	if (infoPtr == 0) {
		char msg[] = "Error getting ssl context info (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
		return 0;
	}
	
	struct ssl_info info;
	long err = bpf_probe_read(&info, sizeof(struct ssl_info), infoPtr);
	
	bpf_map_delete_elem(map_fd, &id);
	
	if (err != 0) {
		char msg[] = "Error reading ssl context (id: %ld) (err: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return 0;
	}
	
	if (info.fd == -1) {
		char msg[] = "File descriptor is missing from ssl info (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
		return 0;
	}
	
	int countBytes = PT_REGS_RC(ctx);
	
	if (info.count_ptr != 0) {
		// ssl_read_ex and ssl_write_ex return 1 for success
		//
		if (countBytes != 1) {
			return 0; 
		}
		
		size_t tempCount;
		long err = bpf_probe_read(&tempCount, sizeof(size_t), (void*) info.count_ptr);
		
		if (err != 0) {
			char msg[] = "Error reading bytes count of _ex (id: %ld) (err: %ld)";
			bpf_trace_printk(msg, sizeof(msg), id, err);
			return 0;
		}
		
		countBytes = tempCount;
	}
	
	if (countBytes <= 0) {
		return 0;
	}
	
	struct tlsChunk* c;
	int zero = 0;
	
	// If other thread, running on the same CPU get to this point at the same time like us
	//	the data will be corrupted - protection may be added in the future
	//	
	c = bpf_map_lookup_elem(&heap, &zero);
	
	if (!c) {
		char msg[] = "Unable to allocate chunk (id: %ld)";
		bpf_trace_printk(msg, sizeof(msg), id);
		return 0;
	}
	
	size_t recorded = MIN(countBytes, sizeof(c->data));
	
	c->flags = flags;
	c->pid = id >> 32;
	c->tgid = id;
	c->len = countBytes;
	c->recorded = recorded;
	c->fd = info.fd;
	
	// This ugly trick is for the ebpf verifier happiness
	//
	if (recorded == sizeof(c->data)) {
		err = bpf_probe_read(c->data, sizeof(c->data), info.buffer);
	} else {
		recorded &= sizeof(c->data) - 1; // Buffer must be N^2
		err = bpf_probe_read(c->data, recorded, info.buffer);
	}
	
	if (err != 0) {
		char msg[] = "Error reading from ssl buffer %ld - %ld";
		bpf_trace_printk(msg, sizeof(msg), id, err);
		return 0;
	}
	
	__u32 pid = id >> 32;
	__u32 fd = info.fd;
	__u64 key = (__u64) pid << 32 | fd;
	
	struct fd_info *fdinfo = bpf_map_lookup_elem(&file_descriptor_to_ipv4, &key);
	
	if (fdinfo != 0) {
		err = bpf_probe_read(c->address, sizeof(c->address), fdinfo->ipv4_addr);
		c->flags |= (fdinfo->flags & FLAGS_IS_CLIENT_BIT);
		
		if (err != 0) {
			char msg[] = "Error reading from fd address %ld - %ld";
			bpf_trace_printk(msg, sizeof(msg), id, err);
		}
	}
	
	bpf_perf_event_output(ctx, &chunks_buffer, BPF_F_CURRENT_CPU, c, sizeof(struct tlsChunk));
	return 0;
}

SEC("uprobe/ssl_write")
int BPF_KPROBE(ssl_write, void* ssl, void* buffer, int num) {
	return ssl_uprobe(ssl, buffer, num, &ssl_write_context, 0);
}

SEC("uretprobe/ssl_write")
int BPF_KPROBE(ssl_ret_write) {
	return ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read")
int BPF_KPROBE(ssl_read, void* ssl, void* buffer, int num) {
	return ssl_uprobe(ssl, buffer, num, &ssl_read_context, 0);
}

SEC("uretprobe/ssl_read")
int BPF_KPROBE(ssl_ret_read) {
	return ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}

SEC("uprobe/ssl_write_ex")
int BPF_KPROBE(ssl_write_ex, void* ssl, void* buffer, size_t num, size_t *written) {
	return ssl_uprobe(ssl, buffer, num, &ssl_write_context, written);
}

SEC("uretprobe/ssl_write_ex")
int BPF_KPROBE(ssl_ret_write_ex) {
	return ssl_uretprobe(ctx, &ssl_write_context, 0);
}

SEC("uprobe/ssl_read_ex")
int BPF_KPROBE(ssl_read_ex, void* ssl, void* buffer, size_t num, size_t *readbytes) {
	return ssl_uprobe(ssl, buffer, num, &ssl_read_context, readbytes);
}

SEC("uretprobe/ssl_read_ex")
int BPF_KPROBE(ssl_ret_read_ex) {
	return ssl_uretprobe(ctx, &ssl_read_context, FLAGS_IS_READ_BIT);
}
