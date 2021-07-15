package mizu

import (
	"regexp"
)

type Version string

func (v Version) Breakdown() (string, string, string) {
	re := regexp.MustCompile(`\d+`)
	breakdown := re.FindAllString(string(v), 3)
	return breakdown[0], breakdown[1], breakdown[2]
}

func (v Version) Major() string {
	major, _, _ := v.Breakdown()
	return major
}

func (v Version) Minor() string {
	_, minor, _ := v.Breakdown()
	return minor
}

func (v Version) Patch() string {
	_, _, patch := v.Breakdown()
	return patch
}
