package main

import (
	"time"
)

type RequiredAcks int16

const (
	RequireNone RequiredAcks = 0
	RequireOne  RequiredAcks = 1
	RequireAll  RequiredAcks = -1
)

func (acks RequiredAcks) String() string {
	switch acks {
	case RequireNone:
		return "none"
	case RequireOne:
		return "one"
	case RequireAll:
		return "all"
	default:
		return "unknown"
	}
}

type UUID struct {
	TimeLow          int32
	TimeMid          int16
	TimeHiAndVersion int16
	ClockSeq         int16
	NodePart1        int32
	NodePart22       int16
}

// Metadata Request (Version: 0)

type MetadataRequestTopicV0 struct {
	Name string
}

type MetadataRequestV0 struct {
	Topics []MetadataRequestTopicV0
}

// Metadata Request (Version: 4)

type MetadataRequestV4 struct {
	Topics                 []MetadataRequestTopicV0
	AllowAutoTopicCreation bool
}

// Metadata Request (Version: 8)

type MetadataRequestV8 struct {
	Topics                             []MetadataRequestTopicV0
	AllowAutoTopicCreation             bool
	IncludeClusterAuthorizedOperations bool
	IncludeTopicAuthorizedOperations   bool
}

// Metadata Request (Version: 10)

type MetadataRequestTopicV10 struct {
	Name string
	UUID UUID
}

type MetadataRequestV10 struct {
	Topics                             []MetadataRequestTopicV10
	AllowAutoTopicCreation             bool
	IncludeClusterAuthorizedOperations bool
	IncludeTopicAuthorizedOperations   bool
}

// Metadata Request (Version: 11)

type MetadataRequestV11 struct {
	Topics                           []MetadataRequestTopicV10
	AllowAutoTopicCreation           bool
	IncludeTopicAuthorizedOperations bool
}

// Metadata Response (Version: 0)

type BrokerV0 struct {
	NodeId int32
	Host   string
	Port   int32
}

type PartitionsV0 struct {
	ErrorCode      int16
	PartitionIndex int32
	LeaderId       int32
	ReplicaNodes   int32
	IsrNodes       int32
}

type TopicV0 struct {
	ErrorCode  int16
	Name       string
	Partitions []PartitionsV0
}

type MetadataResponseV0 struct {
	Brokers []BrokerV0
	Topics  []TopicV0
}

// Metadata Response (Version: 1)

type BrokerV1 struct {
	NodeId int32
	Host   string
	Port   int32
	Rack   string
}

type TopicV1 struct {
	ErrorCode  int16
	Name       string
	IsInternal bool
	Partitions []PartitionsV0
}

type MetadataResponseV1 struct {
	Brokers      []BrokerV1
	ControllerID int32
	Topics       []TopicV1
}

// Metadata Response (Version: 2)

type MetadataResponseV2 struct {
	Brokers      []BrokerV1
	ClusterID    string
	ControllerID int32
	Topics       []TopicV1
}

// Metadata Response (Version: 3)

type MetadataResponseV3 struct {
	ThrottleTimeMs int32
	Brokers        []BrokerV1
	ClusterID      string
	ControllerID   int32
	Topics         []TopicV1
}

// Metadata Response (Version: 5)

type PartitionsV5 struct {
	ErrorCode       int16
	PartitionIndex  int32
	LeaderId        int32
	ReplicaNodes    int32
	IsrNodes        int32
	OfflineReplicas int32
}

type TopicV5 struct {
	ErrorCode  int16
	Name       string
	IsInternal bool
	Partitions []PartitionsV5
}

type MetadataResponseV5 struct {
	ThrottleTimeMs int32
	Brokers        []BrokerV1
	ClusterID      string
	ControllerID   int32
	Topics         []TopicV5
}

// Metadata Response (Version: 7)

type PartitionsV7 struct {
	ErrorCode       int16
	PartitionIndex  int32
	LeaderId        int32
	LeaderEpoch     int32
	ReplicaNodes    int32
	IsrNodes        int32
	OfflineReplicas int32
}

type TopicV7 struct {
	ErrorCode  int16
	Name       string
	IsInternal bool
	Partitions []PartitionsV7
}

type MetadataResponseV7 struct {
	ThrottleTimeMs int32
	Brokers        []BrokerV1
	ClusterID      string
	ControllerID   int32
	Topics         []TopicV7
}

// Metadata Response (Version: 8)

type TopicV8 struct {
	ErrorCode                 int16
	Name                      string
	IsInternal                bool
	Partitions                []PartitionsV7
	TopicAuthorizedOperations int32
}

type MetadataResponseV8 struct {
	ThrottleTimeMs              int32
	Brokers                     []BrokerV1
	ClusterID                   string
	ControllerID                int32
	Topics                      []TopicV8
	ClusterAuthorizedOperations int32
}

// Metadata Response (Version: 10)

type TopicV10 struct {
	ErrorCode                 int16
	Name                      string
	TopicID                   UUID
	IsInternal                bool
	Partitions                []PartitionsV7
	TopicAuthorizedOperations int32
}

type MetadataResponseV10 struct {
	ThrottleTimeMs              int32
	Brokers                     []BrokerV1
	ClusterID                   string
	ControllerID                int32
	Topics                      []TopicV10
	ClusterAuthorizedOperations int32
}

// Metadata Response (Version: 11)

type MetadataResponseV11 struct {
	ThrottleTimeMs int32
	Brokers        []BrokerV1
	ClusterID      string
	ControllerID   int32
	Topics         []TopicV10
}

// ApiVersions Request (Version: 0)

type ApiVersionsRequestV0 struct{}

// ApiVersions Request (Version: 3)

type ApiVersionsRequestV3 struct {
	ClientSoftwareName    string
	ClientSoftwareVersion string
}

// ApiVersions Response (Version: 0)

type ApiVersionsResponseApiKey struct {
	ApiKey     int16
	MinVersion int16
	MaxVersion int16
}

type ApiVersionsResponseV0 struct {
	ErrorCode int16
	ApiKeys   []ApiVersionsResponseApiKey
}

// ApiVersions Response (Version: 1)

type ApiVersionsResponseV1 struct {
	ErrorCode      int16
	ApiKeys        []ApiVersionsResponseApiKey // FIXME: `confluent-kafka-python` causes memory leak
	ThrottleTimeMs int32
}

// Produce Request (Version: 0)

// Message is a kafka message type
type MessageV0 struct {
	Codec            int8        // codec used to compress the message contents
	CompressionLevel int         // compression level
	LogAppendTime    bool        // the used timestamp is LogAppendTime
	Key              []byte      // the message key, may be nil
	Value            []byte      // the message contents
	Set              *MessageSet // the message set a message might wrap
	Version          int8        // v1 requires Kafka 0.10
	Timestamp        time.Time   // the timestamp of the message (version 1+ only)

	compressedSize int // used for computing the compression ratio metrics
}

// MessageBlock represents a part of request with message
type MessageBlock struct {
	Offset int64
	Msg    *MessageV0
}

// MessageSet is a replacement for RecordBatch in older versions
type MessageSet struct {
	PartialTrailingMessage bool // whether the set on the wire contained an incomplete trailing MessageBlock
	OverflowMessage        bool // whether the set on the wire contained an overflow message
	Messages               []*MessageBlock
}

type RecordHeader struct {
	HeaderKeyLength   int8
	HeaderKey         string
	HeaderValueLength int8
	Value             string
}

// Record is kafka record type
type RecordV0 struct {
	Unknown        int8
	Attributes     int8
	TimestampDelta int8
	OffsetDelta    int8
	KeyLength      int8
	Key            string
	ValueLen       int8
	Value          string
	Headers        []RecordHeader
}

// RecordBatch are records from one kafka request
type RecordBatch struct {
	BaseOffset           int64
	BatchLength          int32
	PartitionLeaderEpoch int32
	Magic                int8
	Crc                  int32
	Attributes           int16
	LastOffsetDelta      int32
	FirstTimestamp       int64
	MaxTimestamp         int64
	ProducerId           int64
	ProducerEpoch        int16
	BaseSequence         int32
	Record               []RecordV0
}

type Records struct {
	RecordBatch RecordBatch
	// TODO: Implement `MessageSet`
	// MessageSet  MessageSet
}

type PartitionData struct {
	Index   int32
	Unknown int32
	Records Records
}

type Partitions struct {
	Length        int32
	PartitionData PartitionData
}

type TopicData struct {
	Topic      string
	Partitions Partitions
}

type ProduceRequestV0 struct {
	RequiredAcks RequiredAcks
	Timeout      int32
	TopicData    []TopicData
}

// Produce Request (Version: 3)

type ProduceRequestV3 struct {
	TransactionalID string
	RequiredAcks    RequiredAcks
	Timeout         int32
	TopicData       []TopicData
}

// Produce Response (Version: 0)

type PartitionResponseV0 struct {
	Index      int32
	ErrorCode  int16
	BaseOffset int64
}

type ResponseV0 struct {
	Name               string
	PartitionResponses []PartitionResponseV0
}

type ProduceResponseV0 struct {
	Responses []ResponseV0
}

// Produce Response (Version: 1)

type ProduceResponseV1 struct {
	Responses      []ResponseV0
	ThrottleTimeMs int32
}

// Produce Response (Version: 2)

type PartitionResponseV2 struct {
	Index           int32
	ErrorCode       int16
	BaseOffset      int64
	LogAppendTimeMs int64
}

type ResponseV2 struct {
	Name               string
	PartitionResponses []PartitionResponseV2
}

type ProduceResponseV2 struct {
	Responses      []ResponseV2
	ThrottleTimeMs int32
}

// Produce Response (Version: 5)

type PartitionResponseV5 struct {
	Index           int32
	ErrorCode       int16
	BaseOffset      int64
	LogAppendTimeMs int64
	LogStartOffset  int64
}

type ResponseV5 struct {
	Name               string
	PartitionResponses []PartitionResponseV5
}

type ProduceResponseV5 struct {
	Responses      []ResponseV5
	ThrottleTimeMs int32
}

// Produce Response (Version: 8)

type RecordErrors struct {
	BatchIndex             int32
	BatchIndexErrorMessage string
}

type PartitionResponseV8 struct {
	Index           int32
	ErrorCode       int16
	BaseOffset      int64
	LogAppendTimeMs int64
	LogStartOffset  int64
	RecordErrors    RecordErrors
	ErrorMessage    string
}

type ResponseV8 struct {
	Name               string
	PartitionResponses []PartitionResponseV8
}

type ProduceResponseV8 struct {
	Responses      []ResponseV8
	ThrottleTimeMs int32
}

// Fetch Request (Version: 0)

type FetchPartitionV0 struct {
	Partition         int32
	FetchOffset       int64
	PartitionMaxBytes int32
}

type FetchTopicV0 struct {
	Topic      string
	Partitions []FetchPartitionV0
}

type FetchRequestV0 struct {
	ReplicaId int32
	MaxWaitMs int32
	MinBytes  int32
	Topics    []FetchTopicV0
}

// Fetch Request (Version: 3)

type FetchRequestV3 struct {
	ReplicaId int32
	MaxWaitMs int32
	MinBytes  int32
	MaxBytes  int32
	Topics    []FetchTopicV0
}

// Fetch Request (Version: 4)

type FetchRequestV4 struct {
	ReplicaId      int32
	MaxWaitMs      int32
	MinBytes       int32
	MaxBytes       int32
	IsolationLevel int8
	Topics         []FetchTopicV0
}

// Fetch Request (Version: 5)

type FetchPartitionV5 struct {
	Partition         int32
	FetchOffset       int64
	LogStartOffset    int64
	PartitionMaxBytes int32
}

type FetchTopicV5 struct {
	Topic      string
	Partitions []FetchPartitionV5
}

type FetchRequestV5 struct {
	ReplicaId      int32
	MaxWaitMs      int32
	MinBytes       int32
	MaxBytes       int32
	IsolationLevel int8
	Topics         []FetchTopicV5
}

// Fetch Request (Version: 7)

type ForgottenTopicsDataV7 struct {
	Topic      string
	Partitions []int32
}

type FetchRequestV7 struct {
	ReplicaId           int32
	MaxWaitMs           int32
	MinBytes            int32
	MaxBytes            int32
	IsolationLevel      int8
	SessionId           int32
	SessionEpoch        int32
	Topics              []FetchTopicV5
	ForgottenTopicsData ForgottenTopicsDataV7
}

// Fetch Request (Version: 9)

type FetchPartitionV9 struct {
	Partition          int32
	CurrentLeaderEpoch int32
	FetchOffset        int64
	LogStartOffset     int64
	PartitionMaxBytes  int32
}

type FetchTopicV9 struct {
	Topic      string
	Partitions []FetchPartitionV9
}

type FetchRequestV9 struct {
	ReplicaId           int32
	MaxWaitMs           int32
	MinBytes            int32
	MaxBytes            int32
	IsolationLevel      int8
	SessionId           int32
	SessionEpoch        int32
	Topics              []FetchTopicV9
	ForgottenTopicsData ForgottenTopicsDataV7
}

// Fetch Request (Version: 11)

type FetchRequestV11 struct {
	ReplicaId           int32
	MaxWaitMs           int32
	MinBytes            int32
	MaxBytes            int32
	IsolationLevel      int8
	SessionId           int32
	SessionEpoch        int32
	Topics              []FetchTopicV9
	ForgottenTopicsData ForgottenTopicsDataV7
	RackId              string
}

// Fetch Response (Version: 0)

type PartitionResponseFetchV0 struct {
	Partition     int32
	ErrorCode     int16
	HighWatermark int64
	RecordSet     Records
}

type ResponseFetchV0 struct {
	Topic              string
	PartitionResponses []PartitionResponseFetchV0
}

type FetchResponseV0 struct {
	Responses []ResponseFetchV0
}

// Fetch Response (Version: 1)

type FetchResponseV1 struct {
	ThrottleTimeMs int32
	Responses      []ResponseFetchV0
}

// Fetch Response (Version: 4)

type AbortedTransactionsV4 struct {
	ProducerId  int32
	FirstOffset int32
}

type PartitionResponseFetchV4 struct {
	Partition           int32
	ErrorCode           int16
	HighWatermark       int64
	LastStableOffset    int64
	AbortedTransactions AbortedTransactionsV4
	RecordSet           Records
}

type ResponseFetchV4 struct {
	Topic              string
	PartitionResponses []PartitionResponseFetchV4
}

type FetchResponseV4 struct {
	ThrottleTimeMs int32
	Responses      []ResponseFetchV4
}

// Fetch Response (Version: 5)

type PartitionResponseFetchV5 struct {
	Partition           int32
	ErrorCode           int16
	HighWatermark       int64
	LastStableOffset    int64
	LogStartOffset      int64
	AbortedTransactions AbortedTransactionsV4
	RecordSet           Records
}

type ResponseFetchV5 struct {
	Topic              string
	PartitionResponses []PartitionResponseFetchV5
}

type FetchResponseV5 struct {
	ThrottleTimeMs int32
	Responses      []ResponseFetchV5
}

// Fetch Response (Version: 7)

type FetchResponseV7 struct {
	ThrottleTimeMs int32
	ErrorCode      int16
	SessionId      int32
	Responses      []ResponseFetchV5
}

// Fetch Response (Version: 11)

type PartitionResponseFetchV11 struct {
	Partition            int32
	ErrorCode            int16
	HighWatermark        int64
	LastStableOffset     int64
	LogStartOffset       int64
	AbortedTransactions  AbortedTransactionsV4
	PreferredReadReplica int32
	RecordSet            Records
}

type ResponseFetchV11 struct {
	Topic              string
	PartitionResponses []PartitionResponseFetchV11
}

type FetchResponseV11 struct {
	ThrottleTimeMs int32
	ErrorCode      int16
	SessionId      int32
	Responses      []ResponseFetchV5
}

// ListOffsets Request (Version: 0)

type ListOffsetsRequestPartitionV0 struct {
	PartitionIndex int32
	Timestamp      int64
	MaxNumOffsets  int32
}

type ListOffsetsRequestTopicV0 struct {
	Name       string
	Partitions []ListOffsetsRequestPartitionV0
}

type ListOffsetsRequestV0 struct {
	ReplicaId int32
	Topics    []ListOffsetsRequestTopicV0
}

// ListOffsets Request (Version: 1)

type ListOffsetsRequestPartitionV1 struct {
	PartitionIndex int32
	Timestamp      int64
}

type ListOffsetsRequestTopicV1 struct {
	Name       string
	Partitions []ListOffsetsRequestPartitionV1
}

type ListOffsetsRequestV1 struct {
	ReplicaId int32
	Topics    []ListOffsetsRequestTopicV1
}

// ListOffsets Request (Version: 2)

type ListOffsetsRequestV2 struct {
	ReplicaId      int32
	IsolationLevel int8
	Topics         []ListOffsetsRequestTopicV1
}

// ListOffsets Request (Version: 4)

type ListOffsetsRequestPartitionV4 struct {
	PartitionIndex     int32
	CurrentLeaderEpoch int32
	Timestamp          int64
}

type ListOffsetsRequestTopicV4 struct {
	Name       string
	Partitions []ListOffsetsRequestPartitionV4
}

type ListOffsetsRequestV4 struct {
	ReplicaId int32
	Topics    []ListOffsetsRequestTopicV4
}

// ListOffsets Response (Version: 0)

type ListOffsetsResponsePartitionV0 struct {
	PartitionIndex  int32
	ErrorCode       int16
	OldStyleOffsets int64
}

type ListOffsetsResponseTopicV0 struct {
	Name       string
	Partitions []ListOffsetsResponsePartitionV0
}

type ListOffsetsResponseV0 struct {
	Topics []ListOffsetsResponseTopicV0
}

// ListOffsets Response (Version: 1)

type ListOffsetsResponsePartitionV1 struct {
	PartitionIndex int32
	ErrorCode      int16
	Timestamp      int64
	Offset         int64
}

type ListOffsetsResponseTopicV1 struct {
	Name       string
	Partitions []ListOffsetsResponsePartitionV1
}

type ListOffsetsResponseV1 struct {
	Topics []ListOffsetsResponseTopicV1
}

// ListOffsets Response (Version: 2)

type ListOffsetsResponseV2 struct {
	ThrottleTimeMs int32
	Topics         []ListOffsetsResponseTopicV1
}

// ListOffsets Response (Version: 4)

type ListOffsetsResponsePartitionV4 struct {
	PartitionIndex int32
	ErrorCode      int16
	Timestamp      int64
	Offset         int64
	LeaderEpoch    int32
}

type ListOffsetsResponseTopicV4 struct {
	Name       string
	Partitions []ListOffsetsResponsePartitionV4
}

type ListOffsetsResponseV4 struct {
	Topics []ListOffsetsResponseTopicV4
}

// CreateTopics Request (Version: 0)

type AssignmentsV0 struct {
	PartitionIndex int32
	BrokerIds      []int32
}

type CreateTopicsRequestConfigsV0 struct {
	Name  string
	Value string
}

type CreateTopicsRequestTopicV0 struct {
	Name              string
	NumPartitions     int32
	ReplicationFactor int16
	Assignments       []AssignmentsV0
	Configs           []CreateTopicsRequestConfigsV0
}

type CreateTopicsRequestV0 struct {
	Topics    []CreateTopicsRequestTopicV0
	TimeoutMs int32
}

// CreateTopics Request (Version: 1)

type CreateTopicsRequestV1 struct {
	Topics       []CreateTopicsRequestTopicV0
	TimeoutMs    int32
	ValidateOnly bool
}

// CreateTopics Response (Version: 0)

type CreateTopicsResponseTopicV0 struct {
	Name      string
	ErrorCode int16
}

type CreateTopicsResponseV0 struct {
	Topics []CreateTopicsResponseTopicV0
}

// CreateTopics Response (Version: 1)

type CreateTopicsResponseTopicV1 struct {
	Name         string
	ErrorCode    int16
	ErrorMessage string
}

type CreateTopicsResponseV1 struct {
	Topics []CreateTopicsResponseTopicV1
}

// CreateTopics Response (Version: 2)

type CreateTopicsResponseV2 struct {
	ThrottleTimeMs int32
	Topics         []CreateTopicsResponseTopicV1
}

// CreateTopics Response (Version: 5)

type CreateTopicsResponseConfigsV5 struct {
	Name         string
	Value        string
	ReadOnly     bool
	ConfigSource int8
	IsSensitive  bool
}

type CreateTopicsResponseTopicV5 struct {
	Name              string
	ErrorCode         int16
	ErrorMessage      string
	NumPartitions     int32
	ReplicationFactor int16
	Configs           []CreateTopicsResponseConfigsV5
}

type CreateTopicsResponseV5 struct {
	ThrottleTimeMs int32
	Topics         []CreateTopicsResponseTopicV5
}

// CreateTopics Response (Version: 7)

type CreateTopicsResponseTopicV7 struct {
	Name              string
	TopicID           UUID
	ErrorCode         int16
	ErrorMessage      string
	NumPartitions     int32
	ReplicationFactor int16
	Configs           []CreateTopicsResponseConfigsV5
}

type CreateTopicsResponseV7 struct {
	ThrottleTimeMs int32
	Topics         []CreateTopicsResponseTopicV7
}

// DeleteTopics Request (Version: 0)

type DeleteTopicsRequestV0 struct {
	TopicNames []string
	TimemoutMs int32
}

// DeleteTopics Request (Version: 6)

type DeleteTopicsRequestTopicV6 struct {
	Name string
	UUID UUID
}

type DeleteTopicsRequestV6 struct {
	Topics     []DeleteTopicsRequestTopicV6
	TimemoutMs int32
}

// DeleteTopics Response (Version: 0)

type DeleteTopicsReponseResponseV0 struct {
	Name      string
	ErrorCode int16
}

type DeleteTopicsReponseV0 struct {
	Responses []DeleteTopicsReponseResponseV0
}

// DeleteTopics Response (Version: 1)

type DeleteTopicsReponseV1 struct {
	ThrottleTimeMs int32
	Responses      []DeleteTopicsReponseResponseV0
}

// DeleteTopics Response (Version: 5)

type DeleteTopicsReponseResponseV5 struct {
	Name         string
	ErrorCode    int16
	ErrorMessage string
}

type DeleteTopicsReponseV5 struct {
	ThrottleTimeMs int32
	Responses      []DeleteTopicsReponseResponseV5
}

// DeleteTopics Response (Version: 6)

type DeleteTopicsReponseResponseV6 struct {
	Name         string
	TopicID      UUID
	ErrorCode    int16
	ErrorMessage string
}

type DeleteTopicsReponseV6 struct {
	ThrottleTimeMs int32
	Responses      []DeleteTopicsReponseResponseV6
}
