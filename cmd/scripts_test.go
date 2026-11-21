package cmd

import (
	"testing"

	"github.com/kubeshark/kubeshark/misc"
)

func TestScriptIndex(t *testing.T) {
	scripts := map[int64]misc.ConfigMapScript{
		0: {Title: "alpha"},
		1: {Title: "beta"},
		2: {Title: "gamma"},
	}

	tests := []struct {
		name    string
		scripts map[int64]misc.ConfigMapScript
		title   string
		want    int64
	}{
		{"no scripts yet", map[int64]misc.ConfigMapScript{}, "alpha", 0},
		{"existing title at the lowest key", scripts, "alpha", 0},
		{"existing title in the middle", scripts, "beta", 1},
		{"existing title at the highest key", scripts, "gamma", 2},
		{"new title goes after the highest key", scripts, "delta", 3},
		{"deleted scripts leave a gap", map[int64]misc.ConfigMapScript{0: {Title: "alpha"}, 4: {Title: "beta"}}, "delta", 5},
		{"placeholder title is never matched", map[int64]misc.ConfigMapScript{0: {Title: "New Script"}, 1: {Title: "beta"}}, "New Script", 2},
		{"duplicate titles resolve to the lowest key", map[int64]misc.ConfigMapScript{3: {Title: "alpha"}, 1: {Title: "alpha"}, 2: {Title: "beta"}}, "alpha", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map iteration order is random, so repeat to catch order-dependent results.
			for range 50 {
				if got := scriptIndex(tt.scripts, tt.title); got != tt.want {
					t.Fatalf("scriptIndex(%q) = %d, want %d", tt.title, got, tt.want)
				}
			}
		})
	}
}
