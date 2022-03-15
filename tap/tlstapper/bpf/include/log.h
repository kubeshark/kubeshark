/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __LOG__
#define __LOG__

// The same consts defined in bpf_logger.go
//
#define LOG_LEVEL_ERROR (0)
#define LOG_LEVEL_INFO (1)
#define LOG_LEVEL_DEBUG (2)

// The same struct can be found in bpf_logger.go
//  
//  Be careful when editing, alignment and padding should be exactly the same in go/c.
//
struct log_message {
	__u32 level;
	__u32 message_code;
	__u64 arg1;
	__u64 arg2;
	__u64 arg3;
};

static __always_inline void log_error(void* ctx, __u16 message_code, __u64 arg1, __u64 arg2, __u64 arg3) {
	struct log_message entry = {};
	
	entry.level = LOG_LEVEL_ERROR;
	entry.message_code = message_code;
	entry.arg1 = arg1;
	entry.arg2 = arg2;
	entry.arg3 = arg3;
	
	long err = bpf_perf_event_output(ctx, &log_buffer, BPF_F_CURRENT_CPU, &entry, sizeof(struct log_message));
	
	if (err != 0) {
		char msg[] = "Error writing log error to perf buffer - %ld";
		bpf_trace_printk(msg, sizeof(msg), err);
	}
}

static __always_inline void log_info(void* ctx, __u16 message_code, __u64 arg1, __u64 arg2, __u64 arg3) {
	struct log_message entry = {};
	
	entry.level = LOG_LEVEL_INFO;
	entry.message_code = message_code;
	entry.arg1 = arg1;
	entry.arg2 = arg2;
	entry.arg3 = arg3;
	
	long err = bpf_perf_event_output(ctx, &log_buffer, BPF_F_CURRENT_CPU, &entry, sizeof(struct log_message));
	
	if (err != 0) {
		char msg[] = "Error writing log info to perf buffer - %ld";
		bpf_trace_printk(msg, sizeof(msg), arg1, err);
	}
}

static __always_inline void log_debug(void* ctx, __u16 message_code, __u64 arg1, __u64 arg2, __u64 arg3) {
	struct log_message entry = {};
	
	entry.level = LOG_LEVEL_DEBUG;
	entry.message_code = message_code;
	entry.arg1 = arg1;
	entry.arg2 = arg2;
	entry.arg3 = arg3;
	
	long err = bpf_perf_event_output(ctx, &log_buffer, BPF_F_CURRENT_CPU, &entry, sizeof(struct log_message));
	
	if (err != 0) {
		char msg[] = "Error writing log debug to perf buffer - %ld";
		bpf_trace_printk(msg, sizeof(msg), arg1, err);
	}
}

#endif /* __LOG__ */
