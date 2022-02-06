package version

import (
	"testing"
)

func TestEqualsEquality(t *testing.T) {
	tests := []struct {
		Name   string
		First  string
		Second string
	}{
		{Name: "major", First: "1.0", Second: "1.0"},
		{Name: "patch", First: "1.1", Second: "1.1"},
		{Name: "incremental", First: "1.0-dev0", Second: "1.0-dev0"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if equal, _ := AreEquals(test.First, test.Second); !equal {
				t.Fatalf("Expected %s == %s", test.First, test.Second)
			}
		})
	}
}

func TestEqualsInvalidVersion(t *testing.T) {
	tests := []struct {
		Name   string
		First  string
		Second string
	}{
		{Name: "first semver", First: "1.0.0", Second: "1.0"},
		{Name: "second semver", First: "1.1", Second: "1.1.0"},
		{Name: "incremental invalid", First: "1.0-dev0de", Second: "1.0-dev0"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if _, err := AreEquals(test.First, test.Second); err == nil {
				t.Fatalf("Expected error")
			}
		})
	}
}

func TestEqualsNoEquality(t *testing.T) {
	tests := []struct {
		Name   string
		First  string
		Second string
	}{
		{Name: "major", First: "1.0", Second: "2.0"},
		{Name: "patch", First: "1.0", Second: "1.1"},
		{Name: "incremental", First: "1.0-dev2", Second: "1.0-dev3"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if equal, _ := AreEquals(test.First, test.Second); equal {
				t.Fatalf("Expected %s != %s", test.First, test.Second)
			}
		})
	}
}

func TestGreaterThenGreater(t *testing.T) {
	tests := []struct {
		Name   string
		First  string
		Second string
	}{
		{Name: "major", First: "2.0", Second: "1.0"},
		{Name: "patch", First: "1.1", Second: "1.0"},
		{Name: "incremental", First: "1.0-dev1", Second: "1.0-dev0"},
		{Name: "major vs incremental", First: "1.0", Second: "1.0-dev1"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if greater, _ := GreaterThen(test.First, test.Second); !greater {
				t.Fatalf("Expected %s > %s", test.First, test.Second)
			}
		})
	}
}

func TestGreaterThenLessThen(t *testing.T) {
	tests := []struct {
		Name   string
		First  string
		Second string
	}{
		{Name: "major", First: "1.0", Second: "2.0"},
		{Name: "major equals", First: "1.0", Second: "1.0"},
		{Name: "patch", First: "1.0", Second: "1.1"},
		{Name: "patch equals", First: "1.1", Second: "1.1"},
		{Name: "incremental", First: "1.0-dev0", Second: "1.0-dev1"},
		{Name: "incremental equals", First: "1.0-dev0", Second: "1.0-dev0"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if greater, _ := GreaterThen(test.First, test.Second); greater {
				t.Fatalf("Expected %s < %s", test.First, test.Second)
			}
		})
	}
}
