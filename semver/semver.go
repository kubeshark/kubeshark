package semver

import (
	"regexp"
	"strconv"
)

type SemVersion string

func (v SemVersion) IsValid() bool {
	re := regexp.MustCompile(`\d+`)
	breakdown := re.FindAllString(string(v), 3)

	return len(breakdown) == 3
}

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

func (v SemVersion) GreaterThan(v2 SemVersion) bool {
	vMajor, vMinor, vPatch := v.breakdownInt()
	v2Major, v2Minor, v2Patch := v2.breakdownInt()

	if vMajor > v2Major {
		return true
	} else if vMajor < v2Major {
		return false
	}

	if vMinor > v2Minor {
		return true
	} else if vMinor < v2Minor {
		return false
	}

	if vPatch > v2Patch {
		return true
	}

	return false
}

func (v SemVersion) breakdownInt() (int, int, int) {
	major, minor, patch := v.Breakdown()

	majorInt, _ := strconv.Atoi(major)
	minorInt, _ := strconv.Atoi(minor)
	patchInt, _ := strconv.Atoi(patch)

	return majorInt, minorInt, patchInt
}
