#ifndef __LEGACY_KERNEL_H__
#define __LEGACY_KERNEL_H__

#if defined(bpf_target_x86)

struct thread_struct___v46 {
	struct desc_struct tls_array[3];
	unsigned long sp0;
	unsigned long sp;
	unsigned short es;
	unsigned short ds;
	unsigned short fsindex;
	unsigned short gsindex;
	unsigned long fs;
	unsigned long gs;
	struct perf_event ptrace_bps[4];
	unsigned long debugreg6;
	unsigned long ptrace_dr7;
	unsigned long cr2;
	unsigned long trap_nr;
	unsigned long error_code;
	unsigned long io_bitmap_ptr;
	unsigned long iopl;
	unsigned io_bitmap_max;
    long: 63;
	long: 64;
	long: 64;
	long: 64;
	long: 64;
	long: 64;
	struct fpu fpu;
};

#elif defined(bpf_target_arm)

// Commented out since thread_struct is not used in ARM64 yet.

// struct thread_struct___v46 {
//     struct cpu_context cpu_context;
//     long: 64;
// 	unsigned long tp_value;
// 	struct fpsimd_state fpsimd_state;
// 	unsigned long fault_address;
// 	unsigned long fault_code;
// 	struct debug_info debug;
// };

#endif

#endif /* __LEGACY_KERNEL_H__ */
