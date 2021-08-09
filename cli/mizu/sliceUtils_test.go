package mizu_test

import (
	"github.com/up9inc/mizu/cli/mizu"
	"testing"
)

func TestContainsExists(t *testing.T) {
	tests := []struct {
		slice         []string
		containsValue string
		expected      bool
	}{
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "apple", expected: true},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "orange", expected: true},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "banana", expected: true},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "grapes", expected: true},
	}

	for _, test := range tests {
		actual := mizu.Contains(test.slice, test.containsValue)
		if actual != test.expected {
			t.Errorf("unexpected result - expected: %v, actual: %v", test.expected, actual)
		}
	}
}

func TestContainsNotExists(t *testing.T) {
	tests := []struct {
		slice         []string
		containsValue string
		expected      bool
	}{
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "cat", expected: false},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "dog", expected: false},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "apples", expected: false},
		{slice: []string{"apple", "orange", "banana", "grapes"}, containsValue: "rapes", expected: false},
	}

	for _, test := range tests {
		actual := mizu.Contains(test.slice, test.containsValue)
		if actual != test.expected {
			t.Errorf("unexpected result - expected: %v, actual: %v", test.expected, actual)
		}
	}
}

func TestContainsEmptySlice(t *testing.T) {
	tests := []struct {
		slice         []string
		containsValue string
		expected      bool
	}{
		{slice: []string{}, containsValue: "cat", expected: false},
		{slice: []string{}, containsValue: "dog", expected: false},
	}

	for _, test := range tests {
		actual := mizu.Contains(test.slice, test.containsValue)
		if actual != test.expected {
			t.Errorf("unexpected result - expected: %v, actual: %v", test.expected, actual)
		}
	}
}

func TestContainsNilSlice(t *testing.T) {
	tests := []struct {
		slice         []string
		containsValue string
		expected      bool
	}{
		{slice: nil, containsValue: "cat", expected: false},
		{slice: nil, containsValue: "dog", expected: false},
	}

	for _, test := range tests {
		actual := mizu.Contains(test.slice, test.containsValue)
		if actual != test.expected {
			t.Errorf("unexpected result - expected: %v, actual: %v", test.expected, actual)
		}
	}
}
