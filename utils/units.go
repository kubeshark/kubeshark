package utils

import "github.com/docker/go-units"

func HumanReadableToBytes(humanReadableSize string) (int64, error) {
	return units.FromHumanSize(humanReadableSize)
}
