/*
Note: This file is licenced differently from the rest of the project
SPDX-License-Identifier: GPL-2.0
Copyright (C) UP9 Inc.
*/

#ifndef __PIDS__
#define __PIDS__

int should_tap(__u32 pid) {
	__u32* shouldTap = bpf_map_lookup_elem(&pids_map, &pid);
	
	if (shouldTap != NULL && *shouldTap == 1) {
		return 1;
	}
	
	__u32 globalPid = 0;
	__u32* shouldTapGlobally = bpf_map_lookup_elem(&pids_map, &globalPid);
	
	return shouldTapGlobally != NULL && *shouldTapGlobally == 1;
}

#endif /* __PIDS__ */
