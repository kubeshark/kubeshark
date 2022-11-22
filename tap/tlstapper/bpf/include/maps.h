/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#ifndef __MAPS__
#define __MAPS__

#define FLAGS_IS_CLIENT_BIT (1 << 0)
#define FLAGS_IS_READ_BIT (1 << 1)

#define CHUNK_SIZE (1 << 12)
#define MAX_CHUNKS_PER_OPERATION (8)

// One minute in nano seconds. Chosen by gut feeling.
#define SSL_INFO_MAX_TTL_NANO (1000000000l * 60l)

#define MAX_ENTRIES_HASH        (1 << 12)  // 4096
#define MAX_ENTRIES_PERF_OUTPUT	(1 << 10)  // 1024
#define MAX_ENTRIES_LRU_HASH	(1 << 14)  // 16384

// The same struct can be found in chunk.go
//  
//  Be careful when editing, alignment and padding should be exactly the same in go/c.
//

struct address_info {
    __be32 saddr;
    __be32 daddr;
    __be16 sport;
    __be16 dport;
};

struct tls_chunk {
    __u32 pid;
    __u32 tgid;
    __u32 len;
    __u32 start;
    __u32 recorded;
    __u32 fd;
    __u32 flags;
    struct address_info address_info;
    __u8 data[CHUNK_SIZE]; // Must be N^2
};

struct ssl_info {
    void* buffer;
    __u32 buffer_len;
    __u32 fd;
    __u64 created_at_nano;
    struct address_info address_info;
    
    // for ssl_write and ssl_read must be zero
    // for ssl_write_ex and ssl_read_ex save the *written/*readbytes pointer. 
    //
    size_t *count_ptr;
};

typedef __u8 conn_flags;

struct goid_offsets {
    __u64 g_addr_offset;
    __u64 goid_offset;
};

const struct goid_offsets *unused __attribute__((unused));

// Heap-like area for eBPF programs - stack size limited to 512 bytes, we must use maps for bigger (chunk) objects.
//
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, int);
	__type(value, struct tls_chunk);
} heap SEC(".maps");


#define BPF_MAP(_name, _type, _key_type, _value_type, _max_entries)     \
    struct bpf_map_def SEC("maps") _name = {                            \
        .type = _type,                                                  \
        .key_size = sizeof(_key_type),                                  \
        .value_size = sizeof(_value_type),                              \
        .max_entries = _max_entries,                                    \
    };

#define BPF_HASH(_name, _key_type, _value_type) \
    BPF_MAP(_name, BPF_MAP_TYPE_HASH, _key_type, _value_type, MAX_ENTRIES_HASH)

#define BPF_PERF_OUTPUT(_name) \
    BPF_MAP(_name, BPF_MAP_TYPE_PERF_EVENT_ARRAY, int, __u32, MAX_ENTRIES_PERF_OUTPUT)

#define BPF_LRU_HASH(_name, _key_type, _value_type) \
    BPF_MAP(_name, BPF_MAP_TYPE_LRU_HASH, _key_type, _value_type, MAX_ENTRIES_LRU_HASH)

// Generic
BPF_HASH(pids_map, __u32, __u32);
BPF_LRU_HASH(connection_context, __u64, conn_flags);
BPF_PERF_OUTPUT(chunks_buffer);
BPF_PERF_OUTPUT(log_buffer);

// OpenSSL specific
BPF_LRU_HASH(openssl_write_context, __u64, struct ssl_info);
BPF_LRU_HASH(openssl_read_context, __u64, struct ssl_info);

// Go specific
BPF_HASH(goid_offsets_map, __u32, struct goid_offsets);
BPF_LRU_HASH(go_write_context, __u64, struct ssl_info);
BPF_LRU_HASH(go_read_context, __u64, struct ssl_info);
BPF_LRU_HASH(go_kernel_write_context, __u64, __u32);
BPF_LRU_HASH(go_kernel_read_context, __u64, __u32);
BPF_LRU_HASH(go_user_kernel_write_context, __u64, struct address_info);
BPF_LRU_HASH(go_user_kernel_read_context, __u64, struct address_info);

#endif /* __MAPS__ */
