package har

import "testing"

func TestContentEncoded(t *testing.T) {
	testCases := []struct {
		text        string
		isBinary    bool
		expectedStr string
		binaryLen   int
	}{
		{"not-base64", false, "not-base64", 10},
		{"dGVzdA==", false, "test", 4},
		{"test", true, "\f@A", 3}, // valid UTF-8 with some non-printable chars
		{"IsDggPCAgPiAgID8gICAgN/vv/e/v/u/v7/9v7+/vyIKIu+3kO+3ke+3ku+3k++3lO+3le+3lu+3l++3mO+3me+3mu+3m++3nO+3ne+3nu+3n++3oO+3oe+3ou+3o++3pO+3pe+3pu+3p++3qO+3qe+3qu+3q++3rO+3re+3ru+3ryIK", true, "test", 132}, // invalid UTF-8 (thus binary), taken from https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt
	}

	for _, tc := range testCases {
		c := Content{
			Encoding: "base64",
			Text:     tc.text,
		}
		isBinary, asBytes, asString := c.B64Decoded()
		_ = asBytes

		if tc.isBinary != isBinary {
			t.Errorf("Binary flag mismatch: %t != %t", tc.isBinary, isBinary)
		}

		if !isBinary && tc.expectedStr != asString {
			t.Errorf("Decode value mismatch: %s != %s", tc.expectedStr, asString)
		}

		if tc.binaryLen != len(asBytes) {
			t.Errorf("Binary len mismatch: %d != %d", tc.binaryLen, len(asBytes))
		}
	}
}
