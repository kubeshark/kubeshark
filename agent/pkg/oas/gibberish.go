package oas

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

var (
	patUuid4    = regexp.MustCompile(`(?i)[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	patEmail    = regexp.MustCompile(`^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	patLongNum  = regexp.MustCompile(`^\d{3,}$`)
	patLongNumB = regexp.MustCompile(`[^\d]\d{3,}`)
	patLongNumA = regexp.MustCompile(`\d{3,}[^\d]`)
)

func IsGibberish(str string) bool {
	if IsVersionString(str) {
		return false
	}

	if patEmail.MatchString(str) {
		return true
	}

	if patUuid4.MatchString(str) {
		return true
	}

	if patLongNum.MatchString(str) || patLongNumB.MatchString(str) || patLongNumA.MatchString(str) {
		return true
	}

	//alNum := cleanStr(str, isAlNumRune)
	//alpha := cleanStr(str, isAlphaRune)
	// noiseAll := isNoisy(alNum)
	//triAll := isTrigramBad(strings.ToLower(alpha))
	// _ = noiseAll

	isNotAlNum := func(r rune) bool { return !isAlNumRune(r) }
	chunks := strings.FieldsFunc(str, isNotAlNum)
	noisyLen := 0
	alnumLen := 0
	for _, chunk := range chunks {
		alnumLen += len(chunk)
		noise := isNoisy(chunk)
		tri := isTrigramBad(strings.ToLower(chunk))
		if noise || tri {
			noisyLen += len(chunk)
		}
	}

	return float64(noisyLen) > 0

	//if float64(noisyLen) > 0 {
	//	return true
	//}

	//if len(chunks) > 0 && float64(noisyLen) >= float64(alnumLen)/3.0 {
	//	return true
	//}

	//if triAll {
	//return true
	//}

	// return false
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
				score += 0.10
			case unicode.IsUpper(prev) && unicode.IsDigit(char):
				score += 0.5

			// lower =>
			case unicode.IsLower(prev) && unicode.IsUpper(char):
				score += 0.75
			case unicode.IsLower(prev) && unicode.IsDigit(char):
				score += 0.5

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
	}

	return tgScore, len(trigrams)
}

func isTrigramBad(s string) bool {
	tgScore, cnt := trigramScore(s)

	if cnt > 0 {
		val := math.Sqrt(tgScore) / float64(cnt)
		val2 := tgScore / float64(cnt)
		threshold := 0.005
		bad := val < threshold
		threshold2 := math.Log(float64(cnt)-2) * 0.1
		bad2 := val2 < threshold2
		return bad && bad2 // TODO: improve this logic to be clearer
	}
	return false
}

func isNoisy(s string) bool {
	noise := noiseLevel(s)

	if len(s) > 0 {
		val := (noise * noise) / float64(len(s))
		threshold := 0.6
		bad := val > threshold
		return bad
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
