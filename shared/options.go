package shared

type TrafficFilteringOptions struct {
	IgnoredUserAgents       []string
	PlainTextMaskingRegexes []*SerializableRegexp
	DisableRedaction        bool
}
