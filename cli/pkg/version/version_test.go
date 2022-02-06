package version

import (
	"testing"
)

func TestEqualsEquality(t *testing.T) {
	first, second := "1.0", "1.0"
	if equal, _ := AreEquals(first, second); !equal {
		t.Fatalf("Expected %s == %s", first, second)
	}

	first, second = "1.1", "1.1"
	if equal, _ := AreEquals(first, second); !equal {
		t.Fatalf("Expected %s == %s", first, second)
	}

	first, second = "1.0-dev2", "1.0-dev2"
	if equal, _ := AreEquals(first, second); !equal {
		t.Fatalf("Expected %s == %s", first, second)
	}
}

func TestEqualsNoEquality(t *testing.T) {
	first, second := "1.0", "2.0"
	if equal, _ := AreEquals(first, second); equal {
		t.Fatalf("Expected %s != %s", first, second)
	}

	first, second = "1.0", "1.1"
	if equal, _ := AreEquals(first, second); equal {
		t.Fatalf("Expected %s != %s", first, second)
	}

	first, second = "1.0-dev2", "1.0-dev3"
	if equal, _ := AreEquals(first, second); equal {
		t.Fatalf("Expected %s != %s", first, second)
	}
}

func TestGreaterThenGreater(t *testing.T) {
	first, second := "2.0", "1.0"
	if greater, _ := GreaterThen(first, second); !greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.1", "1.0"
	if greater, _ := GreaterThen(first, second); !greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0-dev1", "1.0-dev0"
	if greater, _ := GreaterThen(first, second); !greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
}

func TestGreaterThenLessthen(t *testing.T) {
	first, second := "1.0", "2.0"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0", "1.0"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0", "1.1"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0", "1.0"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0-dev0", "1.0-dev1"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
	first, second = "1.0-dev0", "1.0-dev0"
	if greater, _ := GreaterThen(first, second); greater {
		t.Fatalf("Expected %s > %s", first, second)
	}
}
