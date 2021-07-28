package providers

import (
	"github.com/patrickmn/go-cache"
	"github.com/up9inc/mizu/shared"
	"time"
)

const tlsLinkRetainmentTime = time.Minute * 15

var (
	TapStatus shared.TapStatus
	RecentTLSLinks = cache.New(tlsLinkRetainmentTime, tlsLinkRetainmentTime)
)
