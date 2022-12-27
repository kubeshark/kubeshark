package utils

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func UnescapeUnicodeCharacters(raw string) string {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(raw), `\\u`, `\u`, -1))
	if err != nil {
		log.Error().Err(err).Send()
		return raw
	}
	return str
}
