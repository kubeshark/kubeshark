/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __COMMON__
#define __COMMON__

const int32_t invalid_fd = -1;

static int get_count_bytes(struct pt_regs *ctx, struct ssl_info* info, __u64 id);
static int add_address_to_chunk(struct pt_regs *ctx, struct tls_chunk* chunk, __u64 id, __u32 fd);
static void send_chunk_part(struct pt_regs *ctx, __u8* buffer, __u64 id, struct tls_chunk* chunk, int start, int end);
static void send_chunk(struct pt_regs *ctx, __u8* buffer, __u64 id, struct tls_chunk* chunk);
static void output_ssl_chunk(struct pt_regs *ctx, struct ssl_info* info, int count_bytes, __u64 id, __u32 flags);
static struct ssl_info new_ssl_info();
static struct ssl_info lookup_ssl_info(struct pt_regs *ctx, struct bpf_map_def* map_fd, __u64 pid_tgid);

#endif /* __COMMON__ */