/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __GOLANG_TYPES__
#define __GOLANG_TYPES__

struct go_interface {
    int64_t type;
    void* ptr;
};

#endif /* __GOLANG_TYPES__ */
