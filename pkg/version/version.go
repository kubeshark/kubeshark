package version

import (
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major       int
	Patch       int
	Incremental int
}

func Parse(ver string) (*Version, error) {
	re := regexp.MustCompile(`^(\d+)\.(\d+)(?:-\w+(\d+))?$`)
	match := re.FindStringSubmatch(ver)
	if len(match) != 4 {
		return nil, fmt.Errorf("invalid format expected <major>.<patch>(-<suffix><incremental>)? %s,", ver)
	}
	major, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing major int: %s, err %w", match[1], err)
	}
	patch, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, fmt.Errorf("error parsing patch int: %s, err %w", match[2], err)
	}

	if match[3] == "" {
		return &Version{Major: major, Patch: patch, Incremental: -1}, nil
	}

	inc, err := strconv.Atoi(match[3])
	if err != nil {
		return nil, fmt.Errorf("Error parsing incremental int: %s, err %w", match[3], err)
	}
	return &Version{Major: major, Patch: patch, Incremental: inc}, nil

}

func AreEquals(first string, second string) (bool, error) {
	firstVer, err := Parse(first)
	if err != nil {
		return false, fmt.Errorf("Failed parsing fist version: %s, error: %w", first, err)
	}
	secondVer, err := Parse(second)
	if err != nil {
		return false, fmt.Errorf("Failed parsing second version: %s, error: %w", second, err)
	}

	return *firstVer == *secondVer, nil
}

func GreaterThen(first string, second string) (bool, error) {
	firstVer, err := Parse(first)
	if err != nil {
		return false, fmt.Errorf("Failed parsing fist version: %s, error: %w", first, err)
	}
	secondVer, err := Parse(second)
	if err != nil {
		return false, fmt.Errorf("Failed parsing second version: %s, error: %w", second, err)
	}

	if firstVer.Major > secondVer.Major {
		return true, nil
	} else if firstVer.Major < secondVer.Major {
		return false, nil
	}

	if firstVer.Patch > secondVer.Patch {
		return true, nil
	} else if firstVer.Patch < secondVer.Patch {
		return false, nil
	}

	if firstVer.Incremental == -1 && secondVer.Incremental > -1 {
		return true, nil
	}

	if firstVer.Incremental > secondVer.Incremental {
		return true, nil
	}
	return false, nil
}
