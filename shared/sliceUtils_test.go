package shared_test

import (
	"fmt"
	"github.com/up9inc/mizu/shared"
	"reflect"
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
			actual := shared.Contains(test.Slice, test.ContainsValue)
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
			actual := shared.Contains(test.Slice, test.ContainsValue)
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
			actual := shared.Contains(test.Slice, test.ContainsValue)
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
			actual := shared.Contains(test.Slice, test.ContainsValue)
			if actual != test.Expected {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}

func TestUniqueNoDuplicateValues(t *testing.T) {
	tests := []struct {
		Slice         []string
		Expected      []string
	}{
		{Slice: []string{"apple", "orange", "banana", "grapes"}, Expected: []string{"apple", "orange", "banana", "grapes"}},
		{Slice: []string{"dog", "cat", "mouse"}, Expected: []string{"dog", "cat", "mouse"}},
	}

	for index, test := range tests {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			actual := shared.Unique(test.Slice)
			if !reflect.DeepEqual(test.Expected, actual) {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}

func TestUniqueDuplicateValues(t *testing.T) {
	tests := []struct {
		Slice         []string
		Expected      []string
	}{
		{Slice: []string{"apple", "apple", "orange", "orange", "banana", "banana", "grapes", "grapes"}, Expected: []string{"apple", "orange", "banana", "grapes"}},
		{Slice: []string{"dog", "cat", "cat", "mouse"}, Expected: []string{"dog", "cat", "mouse"}},
	}

	for index, test := range tests {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			actual := shared.Unique(test.Slice)
			if !reflect.DeepEqual(test.Expected, actual) {
				t.Errorf("unexpected result - Expected: %v, actual: %v", test.Expected, actual)
			}
		})
	}
}
