package semver

import "testing"

func TestGreaterThanComparesNumericComponents(t *testing.T) {
	tests := []struct {
		name string
		v1   SemVersion
		v2   SemVersion
		want bool
	}{
		{
			name: "minor component with fewer digits",
			v1:   "1.16.0",
			v2:   "1.9.0",
			want: true,
		},
		{
			name: "patch component with fewer digits",
			v1:   "1.16.10",
			v2:   "1.16.9",
			want: true,
		},
		{
			name: "lower major component",
			v1:   "1.30.0",
			v2:   "2.0.0",
			want: false,
		},
		{
			name: "same version",
			v1:   "1.16.0",
			v2:   "1.16.0",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.v1.GreaterThan(test.v2); got != test.want {
				t.Fatalf("%s.GreaterThan(%s) = %t, want %t", test.v1, test.v2, got, test.want)
			}
		})
	}
}

func TestGreaterThanSupportsKubernetesGitVersionPrefix(t *testing.T) {
	if !SemVersion("v1.16.0").GreaterThan("v1.9.0") {
		t.Fatal("expected v1.16.0 to be greater than v1.9.0")
	}
}
