package oas

import (
	"testing"
)

func TestAnyJSON(t *testing.T) {
	testCases := []struct {
		inp    string
		isJSON bool
		out    interface{}
	}{
		{`{"key": 1, "keyNull": null}`, true, nil},
		{`[{"key": "val"}, ["subarray"], "string", 1, 2.2, true, null]`, true, nil},
		{`"somestring"`, true, "somestring"},
		{"0", true, 0},
		{"0.5", true, 0.5},
		{"true", true, true},
		{"null", true, nil},
		{"sabbra cadabra", false, nil},
		{"0.1.2.3", false, nil},
	}
	for _, tc := range testCases {
		any, isJSON := anyJSON(tc.inp)
		if isJSON != tc.isJSON {
			t.Errorf("Parse flag mismatch: %t != %t", tc.isJSON, isJSON)
		} else if isJSON && tc.out != nil && tc.out != any {
			t.Errorf("%s != %s", any, tc.out)
		} else if tc.inp == "null" && any != nil {
			t.Errorf("null has to parse as nil (but got %s)", any)
		} else {
			t.Logf("%s => %v", tc.inp, any)
		}
	}
}

func TestStrRunes(t *testing.T) {
	if isAlphaRune('5') {
		t.Logf("Failed")
	}
	if !isAlphaRune('a') {
		t.Logf("Failed")
	}

	if !isAlNumRune('5') {
		t.Logf("Failed")
	}
	if isAlNumRune(' ') {
		t.Logf("Failed")
	}

	if cleanStr("-abc_567", isAlphaRune) != "abc" {
		t.Logf("Failed")
	}

	if cleanStr("-abc_567", isAlNumRune) != "abc567" {
		t.Logf("Failed")
	}
}
