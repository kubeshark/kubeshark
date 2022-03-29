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

// To avoid multiple .o files
//
#include "openssl_uprobes.c"
#include "fd_tracepoints.c"
#include "fd_to_address_tracepoints.c"

char _license[] SEC("license") = "GPL";
