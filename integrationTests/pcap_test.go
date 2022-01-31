package integrationTests

import (
	"testing"
)

func Test(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}
}
