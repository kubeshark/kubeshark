package api

import "regexp"

type SerializableRegexp struct {
	regexp.Regexp
}

func CompileRegexToSerializableRegexp(expr string) (*SerializableRegexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &SerializableRegexp{*re}, nil
}

// UnmarshalText is by json.Unmarshal.
func (r *SerializableRegexp) UnmarshalText(text []byte) error {
	rr, err := CompileRegexToSerializableRegexp(string(text))
	if err != nil {
		return err
	}
	*r = *rr
	return nil
}

// MarshalText is used by json.Marshal.
func (r *SerializableRegexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}