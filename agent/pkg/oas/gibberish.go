package oas

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

var (
	patBase64   = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)
	patUuid4    = regexp.MustCompile(`(?i)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	patEmail    = regexp.MustCompile(`^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	patHexLower = regexp.MustCompile(`(0x)?[0-9a-f]{6,}`)
	patHexUpper = regexp.MustCompile(`(0x)?[0-9A-F]{6,}`)
	patLongNum  = regexp.MustCompile(`^\d{3,}$`)
)

func IsGibberish(str string) bool {
	if IsVersionString(str) {
		return false
	}

	if patBase64.MatchString(str) && len(str) > 32 {
		return true
	}

	if patUuid4.MatchString(str) {
		return true
	}

	if patEmail.MatchString(str) {
		return true
	}

	if patLongNum.MatchString(str) {
		return true
	}

	noAlNum := cleanStr(str, isAlNumRune)
	noiseAll := noiseLevel(noAlNum)
	triAll := isTrigramBad(strings.ToLower(noAlNum))
	_, _ = noiseAll, triAll

	isNotAlNum := func(r rune) bool {
		return !isAlNumRune(r)
	}
	chunks := strings.FieldsFunc(str, isNotAlNum)
	noisyLen := 0
	alnumLen := 0
	for _, chunk := range chunks {
		alnumLen += len(chunk)
		noise := noiseLevel(chunk)
		if noise >= 0.25 || isTrigramBad(strings.ToLower(chunk)) {
			noisyLen += len(chunk)
		}
	}

	if len(chunks) > 0 && float64(noisyLen) >= float64(alnumLen)/3.0 {
		return true
	}

	if triAll {
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
				score += 0.20
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
				score += 1.0

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

func IsVersionString(component string) bool {
	if component == "" {
		return false
	}

	hasV := false
	if strings.HasPrefix(component, "v") {
		component = component[1:]
		hasV = true
	}

	for _, c := range component {
		if string(c) != "." && !unicode.IsDigit(c) {
			return false
		}
	}

	if !hasV && !strings.Contains(component, ".") {
		return false
	}

	return true
}

func trigramScore(str string) (float64, int) {
	tgScore := 0.0
	trigrams := ngrams(str, 3)
	if len(trigrams) > 0 {
		for _, trigram := range trigrams {
			score, found := corpus_trigrams[trigram]
			if found {
				tgScore += score
			}
		}

		tgScore /= float64(len(trigrams))
	}

	return tgScore, len(trigrams)
}

func isTrigramBad(s string) bool {
	tgScore, cnt := trigramScore(s)

	if cnt > 0 {
		threshold := 0.001 * math.Log(float64(cnt))
		return tgScore < threshold
	}
	return false
}

func ngrams(s string, n int) []string {
	result := make([]string, 0)
	for i := 0; i < len(s)-n+1; i++ {
		result = append(result, s[i:i+n])
	}
	return result
}
