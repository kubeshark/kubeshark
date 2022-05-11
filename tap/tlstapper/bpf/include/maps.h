/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __MAPS__
#define __MAPS__

#define FLAGS_IS_CLIENT_BIT (1 << 0)
#define FLAGS_IS_READ_BIT (1 << 1)

#define CHUNK_SIZE (1 << 12)
#define MAX_CHUNKS_PER_OPERATION (8)

// One minute in nano seconds. Chosen by gut feeling.
#define SSL_INFO_MAX_TTL_NANO (1000000000l * 60l)

// The same struct can be found in chunk.go
//  
//  Be careful when editing, alignment and padding should be exactly the same in go/c.
//
struct tlsChunk {
    __u32 pid;
    __u32 tgid;
    __u32 len;
    __u32 start;
    __u32 recorded;
    __u32 fd;
    __u32 flags;
    __u8 address[16];
    __u8 data[CHUNK_SIZE]; // Must be N^2
};

struct ssl_info {
    void* buffer;
    __u32 fd;
    __u64 created_at_nano;
    
    // for ssl_write and ssl_read must be zero
    // for ssl_write_ex and ssl_read_ex save the *written/*readbytes pointer. 
    //
    size_t *count_ptr;
};

struct fd_info {
    __u8 ipv4_addr[16]; // struct sockaddr (linux-src/include/linux/socket.h)
    __u8 flags;
};

#define BPF_MAP(_name, _type, _key_type, _value_type, _max_entries)     \
    struct bpf_map_def SEC("maps") _name = {                            \
        .type = _type,                                                  \
        .key_size = sizeof(_key_type),                                  \
        .value_size = sizeof(_value_type),                              \
        .max_entries = _max_entries,                                    \
    };

#define BPF_HASH(_name, _key_type, _value_type) \
    BPF_MAP(_name, BPF_MAP_TYPE_HASH, _key_type, _value_type, 4096)

#define BPF_PERF_OUTPUT(_name) \
    BPF_MAP(_name, BPF_MAP_TYPE_PERF_EVENT_ARRAY, int, __u32, 1024)
    
#define BPF_LRU_HASH(_name, _key_type, _value_type) \
    BPF_MAP(_name, BPF_MAP_TYPE_LRU_HASH, _key_type, _value_type, 16384)

BPF_HASH(pids_map, __u32, __u32);
BPF_LRU_HASH(ssl_write_context, __u64, struct ssl_info);
BPF_LRU_HASH(ssl_read_context, __u64, struct ssl_info);
BPF_LRU_HASH(file_descriptor_to_ipv4, __u64, struct fd_info);
BPF_PERF_OUTPUT(chunks_buffer);
BPF_PERF_OUTPUT(log_buffer);

#endif /* __MAPS__ */
