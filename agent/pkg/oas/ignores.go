package oas

import "strings"

var ignoredExtensions = []string{"gif", "svg", "css", "png", "ico", "js", "woff2", "woff", "jpg", "jpeg", "swf", "ttf", "map", "webp", "otf", "mp3"}

var ignoredCtypePrefixes = []string{"image/", "font/", "video/", "audio/", "text/javascript"}
var ignoredCtypes = []string{"application/javascript", "application/x-javascript", "text/css", "application/font-woff2", "application/font-woff", "application/x-font-woff"}

var ignoredHeaders = []string{
	"a-im", "accept",
	"authorization", "cache-control", "connection", "content-encoding", "content-length", "content-type", "cookie",
	"date", "dnt", "expect", "forwarded", "from", "front-end-https", "host", "http2-settings",
	"max-forwards", "origin", "pragma", "proxy-authorization", "proxy-connection", "range", "referer",
	"save-data", "te", "trailer", "transfer-encoding", "upgrade", "upgrade-insecure-requests", "x-download-options",
	"server", "user-agent", "via", "warning", "strict-transport-security", "x-permitted-cross-domain-policies",
	"x-att-deviceid", "x-correlation-id", "correlation-id", "x-client-data", "x-dns-prefetch-control",
	"x-http-method-override", "x-real-ip", "x-request-id", "x-request-start", "x-requested-with", "x-uidh",
	"x-same-domain", "x-content-type-options", "x-frame-options", "x-xss-protection",
	"x-wap-profile", "x-scheme", "status", "x-cache", "x-application-context", "retry-after",
	"newrelic", "x-cloud-trace-context", "sentry-trace", "x-cache-hits", "x-served-by", "x-span-name",
	"expires", "set-cookie", "p3p", "content-security-policy", "content-security-policy-report-only",
	"last-modified", "content-language", "x-varnish", "true-client-ip", "akamai-origin-hop",
	"keep-alive", "etag", "alt-svc", "x-csrf-token", "x-ua-compatible", "vary", "x-powered-by",
	"age", "allow", "www-authenticate", "expect-ct", "timing-allow-origin", "referrer-policy",
	"x-aspnet-version", "x-aspnetmvc-version", "x-timer", "x-abuse-info", "x-mod-pagespeed",
	"duration_ms", // UP9 custom
}

var ignoredHeaderPrefixes = []string{
	":", "accept-", "access-control-", "if-", "sec-", "grpc-",
	"x-forwarded-", "x-original-", "cf-",
	"x-up9-", "x-envoy-", "x-hasura-", "x-b3-", "x-datadog-", "x-envoy-", "x-amz-", "x-newrelic-", "x-prometheus-",
	"x-akamai-", "x-spotim-", "x-amzn-", "x-ratelimit-", "x-goog-",
}

func isCtypeIgnored(ctype string) bool {
	for _, prefix := range ignoredCtypePrefixes {
		if strings.HasPrefix(ctype, prefix) {
			return true
		}
	}

	for _, toIgnore := range ignoredCtypes {
		if ctype == toIgnore {
			return true
		}
	}
	return false
}

func isExtIgnored(path string) bool {
	for _, extIgn := range ignoredExtensions {
		if strings.HasSuffix(path, "."+extIgn) {
			return true
		}
	}
	return false
}

func isHeaderIgnored(name string) bool {
	name = strings.ToLower(name)

	for _, ignore := range ignoredHeaders {
		if name == ignore {
			return true
		}
	}

	for _, prefix := range ignoredHeaderPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
