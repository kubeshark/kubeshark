package semver

import (
	"regexp"
)

type SemVersion string

func (v SemVersion) Breakdown() (string, string, string) {
	re := regexp.MustCompile(`\d+`)
	breakdown := re.FindAllString(string(v), 3)
	return breakdown[0], breakdown[1], breakdown[2]
}

func (v SemVersion) Major() string {
	major, _, _ := v.Breakdown()
	return major
}

func (v SemVersion) Minor() string {
	_, minor, _ := v.Breakdown()
	return minor
}

func (v SemVersion) Patch() string {
	_, _, patch := v.Breakdown()
	return patch
}
