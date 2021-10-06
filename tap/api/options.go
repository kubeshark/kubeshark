package api

type TrafficFilteringOptions struct {
	IgnoredUserAgents       []string
	PlainTextMaskingRegexes []*SerializableRegexp
	DisableRedaction        bool
}
