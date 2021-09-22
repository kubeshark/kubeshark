package api

type TrafficFilteringOptions struct {
	HealthChecksUserAgentHeaders []string
	PlainTextMaskingRegexes      []*SerializableRegexp
	DisableRedaction             bool
}
