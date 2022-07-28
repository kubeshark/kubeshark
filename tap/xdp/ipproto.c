// +build ignore

#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/udp.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "../tlstapper/bpf/include/maps.h"


#define MAX_SOCKS 64

static const unsigned char PROTO = IPPROTO_ICMP;

//Ensure map references are available.
/*
        These will be initiated from go and
        referenced in the end BPF opcodes by file descriptor
*/

BPF_ARRAY(xsks_map, int, int, MAX_SOCKS);

BPF_XSK(qidconf_map, int, int, MAX_SOCKS);

SEC("xdp_sock") int xdp_sock_prog(struct xdp_md *ctx)
{

	int *qidconf, index = ctx->rx_queue_index;

	// A set entry here means that the correspnding queue_id
	// has an active AF_XDP socket bound to it.
	qidconf = bpf_map_lookup_elem(&qidconf_map, &index);
	if (qidconf == NULL)
	{
		return XDP_PASS;
	}

	// redirect packets to an xdp socket that match the given IPv4 or IPv6 protocol; pass all other packets to the kernel
	void *data = (void*)(long)ctx->data;
	void *data_end = (void*)(long)ctx->data_end;
	struct ethhdr *eth = data;
	__u16 h_proto = eth->h_proto;
	if ((void*)eth + sizeof(*eth) > data_end) {
		return XDP_PASS;
	}    
		if (bpf_htons(h_proto) == ETH_P_IP) {
			struct iphdr *ip = data + sizeof(*eth);
			if ((void*)ip + sizeof(*ip) <= data_end) {
				if (ip->protocol == PROTO) {
					if (*qidconf != 0)
						return bpf_redirect_map(&xsks_map, index, 0);
				}
			}
		} else if (bpf_htons(h_proto) == ETH_P_IPV6) {
			struct ipv6hdr *ip = data + sizeof(*eth);
			if ((void*)ip + sizeof(*ip) <= data_end) {
				if (ip->nexthdr == PROTO) {
					if (*qidconf)
						return bpf_redirect_map(&xsks_map, index, 0);
				}
			}
		}
	}

	return XDP_PASS;
}

//Basic license just for compiling the object code
char __license[] SEC("license") = "LGPL-2.1 or BSD-2-Clause";
