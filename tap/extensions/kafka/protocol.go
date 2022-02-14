package kafka

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Message is an interface implemented by all request and response types of the
// kafka protocol.
//
// This interface is used mostly as a safe-guard to provide a compile-time check
// for values passed to functions dealing kafka message types.
type Message interface {
	ApiKey() ApiKey
}

type ApiKey int16

func (k ApiKey) String() string {
	if i := int(k); i >= 0 && i < len(apiNames) {
		return apiNames[i]
	}
	return strconv.Itoa(int(k))
}

const (
	// v0  = 0
	v1  = 1
	v2  = 2
	v3  = 3
	v4  = 4
	v5  = 5
	v6  = 6
	v7  = 7
	v8  = 8
	v9  = 9
	v10 = 10
	v11 = 11
)

const (
	Produce                     ApiKey = 0
	Fetch                       ApiKey = 1
	ListOffsets                 ApiKey = 2
	Metadata                    ApiKey = 3
	LeaderAndIsr                ApiKey = 4
	StopReplica                 ApiKey = 5
	UpdateMetadata              ApiKey = 6
	ControlledShutdown          ApiKey = 7
	OffsetCommit                ApiKey = 8
	OffsetFetch                 ApiKey = 9
	FindCoordinator             ApiKey = 10
	JoinGroup                   ApiKey = 11
	Heartbeat                   ApiKey = 12
	LeaveGroup                  ApiKey = 13
	SyncGroup                   ApiKey = 14
	DescribeGroups              ApiKey = 15
	ListGroups                  ApiKey = 16
	SaslHandshake               ApiKey = 17
	ApiVersions                 ApiKey = 18
	CreateTopics                ApiKey = 19
	DeleteTopics                ApiKey = 20
	DeleteRecords               ApiKey = 21
	InitProducerId              ApiKey = 22
	OffsetForLeaderEpoch        ApiKey = 23
	AddPartitionsToTxn          ApiKey = 24
	AddOffsetsToTxn             ApiKey = 25
	EndTxn                      ApiKey = 26
	WriteTxnMarkers             ApiKey = 27
	TxnOffsetCommit             ApiKey = 28
	DescribeAcls                ApiKey = 29
	CreateAcls                  ApiKey = 30
	DeleteAcls                  ApiKey = 31
	DescribeConfigs             ApiKey = 32
	AlterConfigs                ApiKey = 33
	AlterReplicaLogDirs         ApiKey = 34
	DescribeLogDirs             ApiKey = 35
	SaslAuthenticate            ApiKey = 36
	CreatePartitions            ApiKey = 37
	CreateDelegationToken       ApiKey = 38
	RenewDelegationToken        ApiKey = 39
	ExpireDelegationToken       ApiKey = 40
	DescribeDelegationToken     ApiKey = 41
	DeleteGroups                ApiKey = 42
	ElectLeaders                ApiKey = 43
	IncrementalAlterConfigs     ApiKey = 44
	AlterPartitionReassignments ApiKey = 45
	ListPartitionReassignments  ApiKey = 46
	OffsetDelete                ApiKey = 47
	DescribeClientQuotas        ApiKey = 48
	AlterClientQuotas           ApiKey = 49

	numApis = 50
)

var apiNames = [numApis]string{
	Produce:                     "Produce",
	Fetch:                       "Fetch",
	ListOffsets:                 "ListOffsets",
	Metadata:                    "Metadata",
	LeaderAndIsr:                "LeaderAndIsr",
	StopReplica:                 "StopReplica",
	UpdateMetadata:              "UpdateMetadata",
	ControlledShutdown:          "ControlledShutdown",
	OffsetCommit:                "OffsetCommit",
	OffsetFetch:                 "OffsetFetch",
	FindCoordinator:             "FindCoordinator",
	JoinGroup:                   "JoinGroup",
	Heartbeat:                   "Heartbeat",
	LeaveGroup:                  "LeaveGroup",
	SyncGroup:                   "SyncGroup",
	DescribeGroups:              "DescribeGroups",
	ListGroups:                  "ListGroups",
	SaslHandshake:               "SaslHandshake",
	ApiVersions:                 "ApiVersions",
	CreateTopics:                "CreateTopics",
	DeleteTopics:                "DeleteTopics",
	DeleteRecords:               "DeleteRecords",
	InitProducerId:              "InitProducerId",
	OffsetForLeaderEpoch:        "OffsetForLeaderEpoch",
	AddPartitionsToTxn:          "AddPartitionsToTxn",
	AddOffsetsToTxn:             "AddOffsetsToTxn",
	EndTxn:                      "EndTxn",
	WriteTxnMarkers:             "WriteTxnMarkers",
	TxnOffsetCommit:             "TxnOffsetCommit",
	DescribeAcls:                "DescribeAcls",
	CreateAcls:                  "CreateAcls",
	DeleteAcls:                  "DeleteAcls",
	DescribeConfigs:             "DescribeConfigs",
	AlterConfigs:                "AlterConfigs",
	AlterReplicaLogDirs:         "AlterReplicaLogDirs",
	DescribeLogDirs:             "DescribeLogDirs",
	SaslAuthenticate:            "SaslAuthenticate",
	CreatePartitions:            "CreatePartitions",
	CreateDelegationToken:       "CreateDelegationToken",
	RenewDelegationToken:        "RenewDelegationToken",
	ExpireDelegationToken:       "ExpireDelegationToken",
	DescribeDelegationToken:     "DescribeDelegationToken",
	DeleteGroups:                "DeleteGroups",
	ElectLeaders:                "ElectLeaders",
	IncrementalAlterConfigs:     "IncrementalAlterConfigs",
	AlterPartitionReassignments: "AlterPartitionReassignments",
	ListPartitionReassignments:  "ListPartitionReassignments",
	OffsetDelete:                "OffsetDelete",
	DescribeClientQuotas:        "DescribeClientQuotas",
	AlterClientQuotas:           "AlterClientQuotas",
}

type messageType struct {
	version  int16
	flexible bool
	gotype   reflect.Type
	decode   decodeFunc
}

func makeTypes(t reflect.Type) []messageType {
	minVersion := int16(-1)
	maxVersion := int16(-1)

	// All future versions will be flexible (according to spec), so don't need to
	// worry about maxes here.
	minFlexibleVersion := int16(-1)

	forEachStructField(t, func(_ reflect.Type, _ index, tag string) {
		forEachStructTag(tag, func(tag structTag) bool {
			if minVersion < 0 || tag.MinVersion < minVersion {
				minVersion = tag.MinVersion
			}
			if maxVersion < 0 || tag.MaxVersion > maxVersion {
				maxVersion = tag.MaxVersion
			}
			if tag.TagID > -2 && (minFlexibleVersion < 0 || tag.MinVersion < minFlexibleVersion) {
				minFlexibleVersion = tag.MinVersion
			}
			return true
		})
	})

	types := make([]messageType, 0, (maxVersion-minVersion)+1)

	for v := minVersion; v <= maxVersion; v++ {
		flexible := minFlexibleVersion >= 0 && v >= minFlexibleVersion

		types = append(types, messageType{
			version:  v,
			gotype:   t,
			flexible: flexible,
			decode:   decodeFuncOf(t, v, flexible, structTag{}),
		})
	}

	return types
}

type structTag struct {
	MinVersion int16
	MaxVersion int16
	Compact    bool
	Nullable   bool
	TagID      int
}

func forEachStructTag(tag string, do func(structTag) bool) {
	if tag == "-" {
		return // special case to ignore the field
	}

	forEach(tag, '|', func(s string) bool {
		tag := structTag{
			MinVersion: -1,
			MaxVersion: -1,

			// Legitimate tag IDs can start at 0. We use -1 as a placeholder to indicate
			// that the message type is flexible, so that leaves -2 as the default for
			// indicating that there is no tag ID and the message is not flexible.
			TagID: -2,
		}

		var err error
		forEach(s, ',', func(s string) bool {
			switch {
			case strings.HasPrefix(s, "min="):
				tag.MinVersion, err = parseVersion(s[4:])
			case strings.HasPrefix(s, "max="):
				tag.MaxVersion, err = parseVersion(s[4:])
			case s == "tag":
				tag.TagID = -1
			case strings.HasPrefix(s, "tag="):
				tag.TagID, err = strconv.Atoi(s[4:])
			case s == "compact":
				tag.Compact = true
			case s == "nullable":
				tag.Nullable = true
			default:
				err = fmt.Errorf("unrecognized option: %q", s)
			}
			return err == nil
		})

		if err != nil {
			panic(fmt.Errorf("malformed struct tag: %w", err))
		}

		if tag.MinVersion < 0 && tag.MaxVersion >= 0 {
			panic(fmt.Errorf("missing minimum version in struct tag: %q", s))
		}

		if tag.MaxVersion < 0 && tag.MinVersion >= 0 {
			panic(fmt.Errorf("missing maximum version in struct tag: %q", s))
		}

		if tag.MinVersion > tag.MaxVersion {
			panic(fmt.Errorf("invalid version range in struct tag: %q", s))
		}

		return do(tag)
	})
}

func forEach(s string, sep byte, do func(string) bool) bool {
	for len(s) != 0 {
		p := ""
		i := strings.IndexByte(s, sep)
		if i < 0 {
			p, s = s, ""
		} else {
			p, s = s[:i], s[i+1:]
		}
		if !do(p) {
			return false
		}
	}
	return true
}

func forEachStructField(t reflect.Type, do func(reflect.Type, index, string)) {
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)

		if f.PkgPath != "" && f.Name != "_" {
			continue
		}

		kafkaTag, ok := f.Tag.Lookup("kafka")
		if !ok {
			kafkaTag = "|"
		}

		do(f.Type, indexOf(f), kafkaTag)
	}
}

func parseVersion(s string) (int16, error) {
	if !strings.HasPrefix(s, "v") {
		return 0, fmt.Errorf("invalid version number: %q", s)
	}
	i, err := strconv.ParseInt(s[1:], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid version number: %q: %w", s, err)
	}
	if i < 0 {
		return 0, fmt.Errorf("invalid negative version number: %q", s)
	}
	return int16(i), nil
}

func dontExpectEOF(err error) error {
	switch err {
	case nil:
		return nil
	case io.EOF:
		return io.ErrUnexpectedEOF
	default:
		return err
	}
}

type Broker struct {
	ID   int32
	Host string
	Port int32
	Rack string
}

type Topic struct {
	Name       string
	Error      int16
	Partitions map[int32]Partition
}

type Partition struct {
	ID       int32
	Error    int16
	Leader   int32
	Replicas []int32
	ISR      []int32
	Offline  []int32
}

// GroupMessage is an extension of the Message interface implemented by some
// request types to inform the program that they should be routed to a group
// coordinator.
type GroupMessage interface {
	// Returns the group configured on the message.
	Group() string
}

// PreparedMessage is an extension of the Message interface implemented by some
// request types which may need to run some pre-processing on their state before
// being sent.
type PreparedMessage interface {
	// Prepares the message before being sent to a kafka broker using the API
	// version passed as argument.
	Prepare(apiVersion int16)
}

// Merger is an interface implemented by messages which can merge multiple
// results into one response.
type Merger interface {
	// Given a list of message and associated results, merge them back into a
	// response (or an error). The results must be either Message or error
	// values, other types should trigger a panic.
	Merge(messages []Message, results []interface{}) (Message, error)
}
