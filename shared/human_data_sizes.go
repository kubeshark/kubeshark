package shared

import "github.com/docker/go-units"

func BytesToHumanReadable(bytes int64) string {
	return units.HumanSize(float64(bytes))
}

func HumanReadableToBytes(humanReadableSize string) (int64, error) {
	return units.FromHumanSize(humanReadableSize)
}
