package oas

import (
	"regexp"
	"unicode"
)

var (
	patBase64   = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)
	patUuid4    = regexp.MustCompile(`(?i)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	patEmail    = regexp.MustCompile(`^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	patHexLower = regexp.MustCompile(`(0x)?[0-9a-f]{6,}`)
	patHexUpper = regexp.MustCompile(`(0x)?[0-9A-F]{6,}`)
)

func IsGibberish(str string) bool {
	if patBase64.MatchString(str) && len(str) > 32 {
		return true
	}

	if patUuid4.MatchString(str) {
		return true
	}

	if patEmail.MatchString(str) {
		return true
	}

	if patHexLower.MatchString(str) || patHexUpper.MatchString(str) {
		return true
	}

	noise := noiseLevel(str)
	if noise >= 0.2 {
		return true
	}

	return false
}

func noiseLevel(str string) (score float64) {
	// opinionated algo of certain char pairs marking the non-human strings
	prev := *new(rune)
	cnt := 0.0
	for _, char := range str {
		cnt += 1
		if prev > 0 {
			switch {
			// continued class of upper/lower/digit adds no noise
			case unicode.IsUpper(prev) && unicode.IsUpper(char):
			case unicode.IsLower(prev) && unicode.IsLower(char):
			case unicode.IsDigit(prev) && unicode.IsDigit(char):

			// upper =>
			case unicode.IsUpper(prev) && unicode.IsLower(char):
				score += 0.25
			case unicode.IsUpper(prev) && unicode.IsDigit(char):
				score += 0.25

			// lower =>
			case unicode.IsLower(prev) && unicode.IsUpper(char):
				score += 0.75
			case unicode.IsLower(prev) && unicode.IsDigit(char):
				score += 0.25

			// digit =>
			case unicode.IsDigit(prev) && unicode.IsUpper(char):
				score += 0.75
			case unicode.IsDigit(prev) && unicode.IsLower(char):
				score += 0.75

			// the rest is 100% noise
			default:
				score += 1
			}
		}
		prev = char
	}

	score /= cnt // weigh it

	return score
}
