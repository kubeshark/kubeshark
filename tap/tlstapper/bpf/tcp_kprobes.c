#include "include/headers.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"

static __always_inline void tcp_kprobe(struct pt_regs *ctx, struct bpf_map_def *map_fd1, struct bpf_map_def *map_fd2, _Bool is_send) {
	long err;

	__u64 id = bpf_get_current_pid_tgid();
	__u32 pid = id >> 32;

	if (!should_tap(id >> 32)) {
		return;
	}

	_Bool should_forward_to_map = false;
	struct ssl_info *info_ptr = bpf_map_lookup_elem(map_fd1, &id);
	// Happens when the connection is not using openssl lib
	if (info_ptr == NULL) {
		info_ptr = bpf_map_lookup_elem(map_fd2, &id);
		// Happens when the connection is not from a Go program
		if (info_ptr == NULL) {
			return;
		}
		should_forward_to_map = true;
	}

	struct sock *sk = (struct sock *) PT_REGS_PARM1(ctx);

	short unsigned int family;
	err = bpf_probe_read(&family, sizeof(family), (void *)&sk->__sk_common.skc_family);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_FAMILY, id, err, 0l);
		return;
	}
	if (family != AF_INET) {
		return;
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
		return;
	}
	err = bpf_probe_read(&daddr, sizeof(daddr), (void *)&sk->__sk_common.skc_daddr);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_DADDR, id, err, 0l);
		return;
	}
	err = bpf_probe_read(&dport, sizeof(dport), (void *)&sk->__sk_common.skc_dport);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_DPORT, id, err, 0l);
		return;
	}
	err = bpf_probe_read(&sport, sizeof(sport), (void *)&sk->__sk_common.skc_num);
	if (err != 0) {
		log_error(ctx, LOG_ERROR_READING_SOCKET_SPORT, id, err, 0l);
		return;
	}

	if (should_forward_to_map) {
		__u64 key = (__u64) pid << 32 | info_ptr->fd;

		struct fd_info *fdinfo = bpf_map_lookup_elem(&file_descriptor_to_ipv4, &key);
		// Happens when the connection is not tls (there was no connect / accept)
		if (fdinfo == NULL) {
				return;
		}

		fdinfo->address_info.mode = ADDRESS_INFO_MODE_PAIR;
		fdinfo->address_info.daddr = daddr;
		fdinfo->address_info.dport = dport;
		fdinfo->address_info.saddr = saddr;
		fdinfo->address_info.sport = bpf_htons(sport);
	} else {
		info_ptr->address_info.mode = ADDRESS_INFO_MODE_PAIR;
		info_ptr->address_info.daddr = daddr;
		info_ptr->address_info.saddr = saddr;
		info_ptr->address_info.dport = dport;
		info_ptr->address_info.sport = bpf_htons(sport);
	}
}

SEC("kprobe/tcp_sendmsg")
void BPF_KPROBE(tcp_sendmsg) {
	__u64 id = bpf_get_current_pid_tgid();
	tcp_kprobe(ctx, &openssl_write_context, &go_kernel_write_context, true);
}

SEC("kprobe/tcp_recvmsg")
void BPF_KPROBE(tcp_recvmsg) {
	__u64 id = bpf_get_current_pid_tgid();
	tcp_kprobe(ctx, &openssl_read_context, &go_kernel_read_context, false);
}
