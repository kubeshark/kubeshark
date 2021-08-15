package mizu_test

import (
	"github.com/up9inc/mizu/cli/mizu"
	"testing"
)

func TestContainsExists(t *testing.T) {
	tests := []struct {
		Slice         []string
		ContainsValue string
		Expected      bool
	}{
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "apple", Expected: true},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "orange", Expected: true},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "banana", Expected: true},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "grapes", Expected: true},
	}

	for _, test := range tests {
		t.Run(test.ContainsValue, func(t *testing.T) {
			actual := mizu.Contains(test.Slice, test.ContainsValue)
			if actual != test.Expected {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}

func TestContainsNotExists(t *testing.T) {
	tests := []struct {
		Slice         []string
		ContainsValue string
		Expected      bool
	}{
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "cat", Expected: false},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "dog", Expected: false},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "apples", Expected: false},
		{Slice: []string{"apple", "orange", "banana", "grapes"}, ContainsValue: "rapes", Expected: false},
	}

	for _, test := range tests {
		t.Run(test.ContainsValue, func(t *testing.T) {
			actual := mizu.Contains(test.Slice, test.ContainsValue)
			if actual != test.Expected {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}

func TestContainsEmptySlice(t *testing.T) {
	tests := []struct {
		Slice         []string
		ContainsValue string
		Expected      bool
	}{
		{Slice: []string{}, ContainsValue: "cat", Expected: false},
		{Slice: []string{}, ContainsValue: "dog", Expected: false},
	}

	for _, test := range tests {
		t.Run(test.ContainsValue, func(t *testing.T) {
			actual := mizu.Contains(test.Slice, test.ContainsValue)
			if actual != test.Expected {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}

func TestContainsNilSlice(t *testing.T) {
	tests := []struct {
		Slice         []string
		ContainsValue string
		Expected      bool
	}{
		{Slice: nil, ContainsValue: "cat", Expected: false},
		{Slice: nil, ContainsValue: "dog", Expected: false},
	}

	for _, test := range tests {
		t.Run(test.ContainsValue, func(t *testing.T) {
			actual := mizu.Contains(test.Slice, test.ContainsValue)
			if actual != test.Expected {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}
