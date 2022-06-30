#include "include/headers.h"
#include "include/maps.h"
#include "include/log.h"
#include "include/logger_messages.h"
#include "include/pids.h"
#include "include/common.h"

SEC("kprobe/tcp_sendmsg")
void BPF_KPROBE(tcp_sendmsg) {
	long err;

	__u64 id = bpf_get_current_pid_tgid();
	__u32 pid = id >> 32;

	if (!should_tap(id >> 32)) {
		return;
	}

	struct connection_info *connection_info_ptr = bpf_map_lookup_elem(&openssl_connect_context, &id);
	// Happens when the connection is not tls
	if (connection_info_ptr == NULL) {
		return;
	}

	struct sock *sk = (struct sock *) PT_REGS_PARM1(ctx);

	short unsigned int family;
	err = bpf_probe_read(&family, sizeof(family), (void *)&sk->__sk_common.skc_family);
	if (err != 0) {
		// TODO: Raise error
		log_info(ctx, LOG_INFO_DEBUG, -1, 0, 2);
		return;
	}
	if (family != AF_INET) {
		return;
	}

	__be32 saddr;
	__be32 daddr;
	__be16 dport;
	__u16 sport;
	err = bpf_probe_read(&saddr, sizeof(saddr), (void *)&sk->__sk_common.skc_rcv_saddr);
	if (err != 0) {
		// TODO: Raise error
		log_info(ctx, LOG_INFO_DEBUG, -1, 0, 3);
		return;
	}
	err = bpf_probe_read(&daddr, sizeof(daddr), (void *)&sk->__sk_common.skc_daddr);
	if (err != 0) {
		// TODO: Raise error
		log_info(ctx, LOG_INFO_DEBUG, -1, 0, 4);
		return;
	}
	err = bpf_probe_read(&dport, sizeof(dport), (void *)&sk->__sk_common.skc_dport);
	if (err != 0) {
		// TODO: Raise error
		log_info(ctx, LOG_INFO_DEBUG, -1, 0, 5);
		return;
	}
	err = bpf_probe_read(&sport, sizeof(sport), (void *)&sk->__sk_common.skc_num);
	if (err != 0) {
		// TODO: Raise error
		log_info(ctx, LOG_INFO_DEBUG, -1, 0, 6);
		return;
	}

	(void)memcpy(&connection_info_ptr->daddr, &daddr, sizeof(connection_info_ptr->daddr));
	(void)memcpy(&connection_info_ptr->saddr, &saddr, sizeof(connection_info_ptr->saddr));
	connection_info_ptr->dport = dport;
	connection_info_ptr->sport = sport;

	// Debug
	log_info(ctx, LOG_INFO_DEBUG, pid, saddr, daddr);
}
