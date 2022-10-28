/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) Kubeshark
*/

#ifndef __LOG_MESSAGES__
#define __LOG_MESSAGES__

// Must be synced with bpf_logger_messages.go
//
#define LOG_ERROR_READING_BYTES_COUNT (0)
#define LOG_ERROR_READING_FROM_SSL_BUFFER (1)
#define LOG_ERROR_BUFFER_TOO_BIG (2)
#define LOG_ERROR_ALLOCATING_CHUNK (3)
#define LOG_ERROR_READING_SSL_CONTEXT (4)
#define LOG_ERROR_PUTTING_SSL_CONTEXT (5)
#define LOG_ERROR_GETTING_SSL_CONTEXT (6)
#define LOG_ERROR_MISSING_FILE_DESCRIPTOR (7)
#define LOG_ERROR_PUTTING_FILE_DESCRIPTOR (8)
#define LOG_ERROR_PUTTING_ACCEPT_INFO (9)
#define LOG_ERROR_GETTING_ACCEPT_INFO (10)
#define LOG_ERROR_READING_ACCEPT_INFO (11)
#define LOG_ERROR_PUTTING_CONNECTION_CONTEXT (12)
#define LOG_ERROR_PUTTING_CONNECT_INFO (13)
#define LOG_ERROR_GETTING_CONNECT_INFO (14)
#define LOG_ERROR_READING_CONNECT_INFO (15)
#define LOG_ERROR_READING_SOCKET_FAMILY (16)
#define LOG_ERROR_READING_SOCKET_DADDR (17)
#define LOG_ERROR_READING_SOCKET_SADDR (18)
#define LOG_ERROR_READING_SOCKET_DPORT (19)
#define LOG_ERROR_READING_SOCKET_SPORT (20)
#define LOG_ERROR_PUTTING_GO_USER_KERNEL_CONTEXT (21)
#define LOG_ERROR_GETTING_GO_USER_KERNEL_CONTEXT (22)

// Sometimes we have the same error, happening from different locations.
// 	in order to be able to distinct between them in the log, we add an 
// 	extra number that identify the location. The number can be anything, 
//	but do not give the same number to different origins.
// 
#define ORIGIN_SSL_UPROBE_CODE (0l)
#define ORIGIN_SSL_URETPROBE_CODE (1l)
#define ORIGIN_SYS_ENTER_READ_CODE (2l)
#define ORIGIN_SYS_ENTER_WRITE_CODE (3l)
#define ORIGIN_SYS_EXIT_ACCEPT4_CODE (4l)
#define ORIGIN_SYS_EXIT_CONNECT_CODE (5l)

#endif /* __LOG_MESSAGES__ */
