package providers

import (
	"github.com/patrickmn/go-cache"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"time"
)

const tlsLinkRetainmentTime = time.Minute * 15

var (
	TapStatus shared.TapStatus
	RecentTLSLinks = cache.New(tlsLinkRetainmentTime, tlsLinkRetainmentTime)
)

func GetAllRecentTLSAddresses() []string {
	recentTLSLinks := make([]string, 0)

	for _, outboundLinkItem := range RecentTLSLinks.Items() {
		outboundLink, castOk := outboundLinkItem.Object.(*tap.OutboundLink)
		if castOk {
			recentTLSLinks = append(recentTLSLinks, outboundLink.DstIP)
		}
	}

	return recentTLSLinks
}
