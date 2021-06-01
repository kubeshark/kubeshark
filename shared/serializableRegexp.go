package shared

import "regexp"

type SerializableRegexp struct {
	regexp.Regexp
}

// CompileRegexToSerializableRegexp wraps the result of the standard library's
// regexp.Compile, for easy (un)marshaling.
func CompileRegexToSerializableRegexp(expr string) (*SerializableRegexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &SerializableRegexp{*re}, nil
}

// UnmarshalText satisfies the encoding.TextMarshaler interface,
// also used by json.Unmarshal.
func (r *SerializableRegexp) UnmarshalText(text []byte) error {
	rr, err := CompileRegexToSerializableRegexp(string(text))
	if err != nil {
		return err
	}
	*r = *rr
	return nil
}

// MarshalText satisfies the encoding.TextMarshaler interface,
// also used by json.Marshal.
func (r *SerializableRegexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}
