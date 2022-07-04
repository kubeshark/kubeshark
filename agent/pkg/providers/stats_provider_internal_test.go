package providers

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestGetBucketOfTimeStamp(t *testing.T) {
	tests := map[int64]time.Time{
		time.Date(2022, time.Month(1), 1, 10, 34, 45, 0, time.Local).UnixMilli(): time.Date(2022, time.Month(1), 1, 10, 34, 00, 0, time.Local),
		time.Date(2022, time.Month(1), 1, 10, 34, 00, 0, time.Local).UnixMilli(): time.Date(2022, time.Month(1), 1, 10, 34, 00, 0, time.Local),
		time.Date(2022, time.Month(1), 1, 10, 59, 01, 0, time.Local).UnixMilli(): time.Date(2022, time.Month(1), 1, 10, 59, 00, 0, time.Local),
	}

	for key, value := range tests {
		t.Run(fmt.Sprintf("%v", key), func(t *testing.T) {

			actual := getBucketFromTimeStamp(key)

			if actual != value {
				t.Errorf("unexpected result - expected: %v, actual: %v", value, actual)
			}
		})
	}
}

func TestGetAggregatedStatsAllTime(t *testing.T) {
	bucketStatsForTest := BucketStats{
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 00, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
						"post": {
							EntriesCount:  2,
							VolumeInBytes: 3,
						},
					},
				},
				"kafka": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"listTopics": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 01, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
						"post": {
							EntriesCount:  2,
							VolumeInBytes: 3,
						},
					},
				},
				"redis": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"set": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
	}

	expected := map[string]map[string]*AccumulativeStatsCounter{
		"http": {
			"post": {
				Name:            "post",
				EntriesCount:    4,
				VolumeSizeBytes: 6,
			},
			"get": {
				Name:            "get",
				EntriesCount:    2,
				VolumeSizeBytes: 4,
			},
		},
		"kafka": {
			"listTopics": {
				Name:            "listTopics",
				EntriesCount:    5,
				VolumeSizeBytes: 6,
			},
		},
		"redis": {
			"set": {
				Name:            "set",
				EntriesCount:    5,
				VolumeSizeBytes: 6,
			},
		},
	}
	actual := getAggregatedStats(bucketStatsForTest)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected result - expected: %v, actual: %v", 3, len(actual))
	}
}

func TestGetAggregatedStatsFromSpecificTime(t *testing.T) {
	bucketStatsForTest := BucketStats{
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 00, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
					},
				},
				"kafka": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"listTopics": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 01, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
						"post": {
							EntriesCount:  2,
							VolumeInBytes: 3,
						},
					},
				},
				"redis": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"set": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
	}

	expected := map[time.Time]map[string]map[string]*AccumulativeStatsCounter{
		time.Date(2022, time.Month(1), 1, 10, 00, 00, 0, time.UTC): {
			"http": {
				"post": {
					Name:            "post",
					EntriesCount:    2,
					VolumeSizeBytes: 3,
				},
				"get": {
					Name:            "get",
					EntriesCount:    2,
					VolumeSizeBytes: 4,
				},
			},
			"kafka": {
				"listTopics": {
					Name:            "listTopics",
					EntriesCount:    5,
					VolumeSizeBytes: 6,
				},
			},
			"redis": {
				"set": {
					Name:            "set",
					EntriesCount:    5,
					VolumeSizeBytes: 6,
				},
			},
		},
	}
	actual := getAggregatedResultTiming(time.Minute*5, bucketStatsForTest)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected result - expected: %v, actual: %v", 3, len(actual))
	}
}

func TestGetAggregatedStatsFromSpecificTimeMultipleBuckets(t *testing.T) {
	bucketStatsForTest := BucketStats{
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 00, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
					},
				},
				"kafka": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"listTopics": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
		&TimeFrameStatsValue{
			BucketTime: time.Date(2022, time.Month(1), 1, 10, 01, 00, 0, time.UTC),
			ProtocolStats: map[string]ProtocolStats{
				"http": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"get": {
							EntriesCount:  1,
							VolumeInBytes: 2,
						},
						"post": {
							EntriesCount:  2,
							VolumeInBytes: 3,
						},
					},
				},
				"redis": {
					MethodsStats: map[string]*SizeAndEntriesCount{
						"set": {
							EntriesCount:  5,
							VolumeInBytes: 6,
						},
					},
				},
			},
		},
	}

	expected := map[time.Time]map[string]map[string]*AccumulativeStatsCounter{
		time.Date(2022, time.Month(1), 1, 10, 00, 00, 0, time.UTC): {
			"http": {
				"get": {
					Name:            "get",
					EntriesCount:    1,
					VolumeSizeBytes: 2,
				},
			},
			"kafka": {
				"listTopics": {
					Name:            "listTopics",
					EntriesCount:    5,
					VolumeSizeBytes: 6,
				},
			},
		},
		time.Date(2022, time.Month(1), 1, 10, 01, 00, 0, time.UTC): {
			"http": {
				"post": {
					Name:            "post",
					EntriesCount:    2,
					VolumeSizeBytes: 3,
				},
				"get": {
					Name:            "get",
					EntriesCount:    1,
					VolumeSizeBytes: 2,
				},
			},
			"redis": {
				"set": {
					Name:            "set",
					EntriesCount:    5,
					VolumeSizeBytes: 6,
				},
			},
		},
	}
	actual := getAggregatedResultTiming(time.Minute, bucketStatsForTest)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected result - expected: %v, actual: %v", 3, len(actual))
	}
}
