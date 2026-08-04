package utils

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func UnescapeUnicodeCharacters(raw string) string {
	str, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(raw), `\\u`, `\u`))
	if err != nil {
		log.Error().Err(err).Send()
		return raw
	}
	return str
}
