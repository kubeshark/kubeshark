#include "include/headers.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"


static __always_inline int tcp_kprobes_get_address_pair_from_ctx(struct pt_regs *ctx, __u64 id, struct address_info *address_info_ptr) {
	long err;
	struct sock *sk = (struct sock *) PT_REGS_PARM1(ctx);

	short unsigned int family;
	err = bpf_probe_read(&family, sizeof(family), (void *)&sk->__sk_common.skc_family);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_FAMILY, id, err, 0l);
		return -1;
	}
	if (family != AF_INET) {
		return -1;
	}

	// daddr, saddr and dport are in network byte order (big endian)
	// sport is in host byte order
	__be32 saddr;
	__be32 daddr;
	__be16 dport;
	__u16 sport;

	err = bpf_probe_read(&saddr, sizeof(saddr), (void *)&sk->__sk_common.skc_rcv_saddr);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_SADDR, id, err, 0l);
		return -1;
	}
	err = bpf_probe_read(&daddr, sizeof(daddr), (void *)&sk->__sk_common.skc_daddr);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_DADDR, id, err, 0l);
		return -1;
	}
	err = bpf_probe_read(&dport, sizeof(dport), (void *)&sk->__sk_common.skc_dport);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_DPORT, id, err, 0l);
		return -1;
	}
	err = bpf_probe_read(&sport, sizeof(sport), (void *)&sk->__sk_common.skc_num);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_SPORT, id, err, 0l);
		return -1;
	}

	address_info_ptr->daddr = daddr;
	address_info_ptr->saddr = saddr;
	address_info_ptr->dport = dport;
	address_info_ptr->sport = bpf_htons(sport);

	return 0;
}

static __always_inline void tcp_kprobes_forward_go(struct pt_regs *ctx, __u64 id, __u32 fd, struct address_info address_info, struct bpf_map_def *map_fd_go_user_kernel) {
		__u32 pid = id >> 32;
		__u64 key = (__u64) pid << 32 | fd;

		long err = bpf_map_update_elem(map_fd_go_user_kernel, &key, &address_info, BPF_ANY);
    if (err != 0) {
        log_error(ctx, LOG_ERROR_PUTTING_GO_USER_KERNEL_CONTEXT, id, fd, err);
				return;
    }
}

static void __always_inline tcp_kprobes_forward_openssl(struct ssl_info *info_ptr, struct address_info address_info) {
		info_ptr->address_info.daddr = address_info.daddr;
		info_ptr->address_info.saddr = address_info.saddr;
		info_ptr->address_info.dport = address_info.dport;
		info_ptr->address_info.sport = address_info.sport;
}

static __always_inline void tcp_kprobe(struct pt_regs *ctx, struct bpf_map_def *map_fd_openssl, struct bpf_map_def *map_fd_go_kernel, struct bpf_map_def *map_fd_go_user_kernel) {
	long err;

	__u64 id = bpf_get_current_pid_tgid();

	if (!should_tap(id >> 32)) {
		return;
	}

	struct address_info address_info;
	if (0 != tcp_kprobes_get_address_pair_from_ctx(ctx, id, &address_info)) {
		return;
	}

	struct ssl_info *info_ptr = bpf_map_lookup_elem(map_fd_openssl, &id);
	__u32 *fd_ptr;
	if (info_ptr == NULL) {
		fd_ptr = bpf_map_lookup_elem(map_fd_go_kernel, &id);
		// Connection is used by a Go program
		if (fd_ptr == NULL) {
			// Connection was not created by a Go program or by openssl lib
			return;
		}
		tcp_kprobes_forward_go(ctx, id, *fd_ptr, address_info, map_fd_go_user_kernel);
	} else {
		// Connection is used by openssl lib
		tcp_kprobes_forward_openssl(info_ptr, address_info);
	}

}

SEC("kprobe/tcp_sendmsg")
void BPF_KPROBE(tcp_sendmsg) {
	__u64 id = bpf_get_current_pid_tgid();
	tcp_kprobe(ctx, &openssl_write_context, &go_kernel_write_context, &go_user_kernel_write_context);
}

SEC("kprobe/tcp_recvmsg")
void BPF_KPROBE(tcp_recvmsg) {
	__u64 id = bpf_get_current_pid_tgid();
	tcp_kprobe(ctx, &openssl_read_context, &go_kernel_read_context, &go_user_kernel_read_context);
}
