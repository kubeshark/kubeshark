package oas

import (
	"regexp"
)

var (
	patBase64 = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)
	patUuid4  = regexp.MustCompile(`(?i)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	patEmail  = regexp.MustCompile(`^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	patHex    = regexp.MustCompile(`(0x)?[0-9a-f]{6,}`)
)

func isGibberish(str string) bool {
	if patBase64.MatchString(str) && len(str) > 16 {
		return true
	}

	if patUuid4.MatchString(str) {
		return true
	}

	if patEmail.MatchString(str) {
		return true
	}

	if patHex.MatchString(str) {
		return true
	}

	return false
}
