package tap

import (
	"net"
	"strings"

	"github.com/up9inc/mizu/tap/diagnose"
)

var privateIPBlocks []*net.IPNet

func init() {
	initPrivateIPBlocks()
}

// Get this host ipv4 and ipv6 addresses on all interfaces
func getLocalhostIPs() ([]string, error) {
	addrMasks, err := net.InterfaceAddrs()
	if err != nil {
		// TODO: return error, log error
		return nil, err
	}

	myIPs := make([]string, len(addrMasks))
	for ii, addr := range addrMasks {
		myIPs[ii] = strings.Split(addr.String(), "/")[0]
	}

	return myIPs, nil
}

func initPrivateIPBlocks() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			diagnose.TapErrors.Error("Private-IP-Block-Parse", "parse error on %q: %v", cidr, err)
		} else {
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}
}
