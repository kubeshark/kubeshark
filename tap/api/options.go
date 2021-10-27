package api

import "github.com/up9inc/mizu/shared"

type TrafficFilteringOptions struct {
	IgnoredUserAgents       []string
	PlainTextMaskingRegexes []*shared.SerializableRegexp
	DisableRedaction        bool
}
