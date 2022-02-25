package oas

import (
	"encoding/json"
	"github.com/chanced/openapi"
	"reflect"
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
			t.Logf("%s => %s", tc.inp, any)
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

func TestOpMerging(t *testing.T) {
	testCases := []struct {
		op1 *openapi.Operation
		op2 *openapi.Operation
		res *openapi.Operation
	}{
		{nil, nil, nil},
		{&openapi.Operation{}, nil, &openapi.Operation{}},
		{nil, &openapi.Operation{}, &openapi.Operation{}},
		{
			&openapi.Operation{OperationID: "op1"},
			&openapi.Operation{OperationID: "op2"},
			&openapi.Operation{OperationID: "op1", Extensions: openapi.Extensions{}},
		},
		// has historicIds
	}
	for _, tc := range testCases {
		mergeOps(&tc.op1, &tc.op2)

		if !reflect.DeepEqual(tc.op1, tc.res) {
			txt, _ := json.Marshal(tc.op1)
			t.Errorf("Does not match expected: %s", txt)
		}
	}
}
