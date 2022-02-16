package kafka

import (
	"time"
)

type RequiredAcks int16

const (
	RequireNone RequiredAcks = 0
	RequireOne  RequiredAcks = 1
	RequireAll  RequiredAcks = -1
)

type UUID struct {
	TimeLow          int32 `json:"timeLow"`
	TimeMid          int16 `json:"timeMid"`
	TimeHiAndVersion int16 `json:"timeHiAndVersion"`
	ClockSeq         int16 `json:"clockSeq"`
	NodePart1        int32 `json:"nodePart1"`
	NodePart22       int16 `json:"nodePart22"`
}

// Metadata Request (Version: 0)

type MetadataRequestTopicV0 struct {
	Name string `json:"name"`
}

type MetadataRequestV0 struct {
	Topics []MetadataRequestTopicV0 `json:"topics"`
}

// Metadata Request (Version: 4)

type MetadataRequestV4 struct {
	Topics                 []MetadataRequestTopicV0 `json:"topics"`
	AllowAutoTopicCreation bool                     `json:"allowAutoTopicCreation"`
}

// Metadata Request (Version: 8)

type MetadataRequestV8 struct {
	Topics                             []MetadataRequestTopicV0 `json:"topics"`
	AllowAutoTopicCreation             bool                     `json:"allowAutoTopicCreation"`
	IncludeClusterAuthorizedOperations bool                     `json:"includeClusterAuthorizedOperations"`
	IncludeTopicAuthorizedOperations   bool                     `json:"includeTopicAuthorizedOperations"`
}

// Metadata Request (Version: 10)

type MetadataRequestTopicV10 struct {
	Name string
	UUID UUID
}

type MetadataRequestV10 struct {
	Topics                             []MetadataRequestTopicV10 `json:"topics"`
	AllowAutoTopicCreation             bool                      `json:"allowAutoTopicCreation"`
	IncludeClusterAuthorizedOperations bool                      `json:"includeClusterAuthorizedOperations"`
	IncludeTopicAuthorizedOperations   bool                      `json:"includeTopicAuthorizedOperations"`
}

// Metadata Request (Version: 11)

type MetadataRequestV11 struct {
	Topics                           []MetadataRequestTopicV10 `json:"topics"`
	AllowAutoTopicCreation           bool                      `json:"allowAutoTopicCreation"`
	IncludeTopicAuthorizedOperations bool                      `json:"includeTopicAuthorizedOperations"`
}

// Metadata Response (Version: 0)

type BrokerV0 struct {
	NodeId int32  `json:"nodeId"`
	Host   string `json:"host"`
	Port   int32  `json:"port"`
}

type PartitionsV0 struct {
	ErrorCode      int16 `json:"errorCode"`
	PartitionIndex int32 `json:"partitionIndex"`
	LeaderId       int32 `json:"leaderId"`
	ReplicaNodes   int32 `json:"replicaNodes"`
	IsrNodes       int32 `json:"isrNodes"`
}

type TopicV0 struct {
	ErrorCode  int16          `json:"errorCode"`
	Name       string         `json:"name"`
	Partitions []PartitionsV0 `json:"partitions"`
}

type MetadataResponseV0 struct {
	Brokers []BrokerV0 `json:"brokers"`
	Topics  []TopicV0  `json:"topics"`
}

// Metadata Response (Version: 1)

type BrokerV1 struct {
	NodeId int32  `json:"nodeId"`
	Host   string `json:"host"`
	Port   int32  `json:"port"`
	Rack   string `json:"rack"`
}

type TopicV1 struct {
	ErrorCode  int16          `json:"errorCode"`
	Name       string         `json:"name"`
	IsInternal bool           `json:"isInternal"`
	Partitions []PartitionsV0 `json:"partitions"`
}

type MetadataResponseV1 struct {
	Brokers      []BrokerV1 `json:"brokers"`
	ControllerID int32      `json:"controllerID"`
	Topics       []TopicV1  `json:"topics"`
}

// Metadata Response (Version: 2)

type MetadataResponseV2 struct {
	Brokers      []BrokerV1 `json:"brokers"`
	ClusterID    string     `json:"clusterID"`
	ControllerID int32      `json:"controllerID"`
	Topics       []TopicV1  `json:"topics"`
}

// Metadata Response (Version: 3)

type MetadataResponseV3 struct {
	ThrottleTimeMs int32      `json:"throttleTimeMs"`
	Brokers        []BrokerV1 `json:"brokers"`
	ClusterID      string     `json:"clusterID"`
	ControllerID   int32      `json:"controllerID"`
	Topics         []TopicV1  `json:"topics"`
}

// Metadata Response (Version: 5)

type PartitionsV5 struct {
	ErrorCode       int16 `json:"errorCode"`
	PartitionIndex  int32 `json:"partitionIndex"`
	LeaderId        int32 `json:"leaderId"`
	ReplicaNodes    int32 `json:"replicaNodes"`
	IsrNodes        int32 `json:"isrNodes"`
	OfflineReplicas int32 `json:"offlineReplicas"`
}

type TopicV5 struct {
	ErrorCode  int16          `json:"errorCode"`
	Name       string         `json:"name"`
	IsInternal bool           `json:"isInternal"`
	Partitions []PartitionsV5 `json:"partitions"`
}

type MetadataResponseV5 struct {
	ThrottleTimeMs int32      `json:"throttleTimeMs"`
	Brokers        []BrokerV1 `json:"brokers"`
	ClusterID      string     `json:"clusterID"`
	ControllerID   int32      `json:"controllerID"`
	Topics         []TopicV5  `json:"topics"`
}

// Metadata Response (Version: 7)

type PartitionsV7 struct {
	ErrorCode       int16 `json:"errorCode"`
	PartitionIndex  int32 `json:"partitionIndex"`
	LeaderId        int32 `json:"leaderId"`
	LeaderEpoch     int32 `json:"leaderEpoch"`
	ReplicaNodes    int32 `json:"replicaNodes"`
	IsrNodes        int32 `json:"isrNodes"`
	OfflineReplicas int32 `json:"offlineReplicas"`
}

type TopicV7 struct {
	ErrorCode  int16          `json:"errorCode"`
	Name       string         `json:"name"`
	IsInternal bool           `json:"isInternal"`
	Partitions []PartitionsV7 `json:"partitions"`
}

type MetadataResponseV7 struct {
	ThrottleTimeMs int32      `json:"throttleTimeMs"`
	Brokers        []BrokerV1 `json:"brokers"`
	ClusterID      string     `json:"clusterID"`
	ControllerID   int32      `json:"controllerID"`
	Topics         []TopicV7  `json:"topics"`
}

// Metadata Response (Version: 8)

type TopicV8 struct {
	ErrorCode                 int16          `json:"errorCode"`
	Name                      string         `json:"name"`
	IsInternal                bool           `json:"isInternal"`
	Partitions                []PartitionsV7 `json:"partitions"`
	TopicAuthorizedOperations int32          `json:"topicAuthorizedOperations"`
}

type MetadataResponseV8 struct {
	ThrottleTimeMs              int32      `json:"throttleTimeMs"`
	Brokers                     []BrokerV1 `json:"brokers"`
	ClusterID                   string     `json:"clusterID"`
	ControllerID                int32      `json:"controllerID"`
	Topics                      []TopicV8  `json:"topics"`
	ClusterAuthorizedOperations int32      `json:"clusterAuthorizedOperations"`
}

// Metadata Response (Version: 10)

type TopicV10 struct {
	ErrorCode                 int16          `json:"errorCode"`
	Name                      string         `json:"name"`
	TopicID                   UUID           `json:"topicID"`
	IsInternal                bool           `json:"isInternal"`
	Partitions                []PartitionsV7 `json:"partitions"`
	TopicAuthorizedOperations int32          `json:"topicAuthorizedOperations"`
}

type MetadataResponseV10 struct {
	ThrottleTimeMs              int32      `json:"throttleTimeMs"`
	Brokers                     []BrokerV1 `json:"brokers"`
	ClusterID                   string     `json:"clusterID"`
	ControllerID                int32      `json:"controllerID"`
	Topics                      []TopicV10 `json:"topics"`
	ClusterAuthorizedOperations int32      `json:"clusterAuthorizedOperations"`
}

// Metadata Response (Version: 11)

type MetadataResponseV11 struct {
	ThrottleTimeMs int32      `json:"throttleTimeMs"`
	Brokers        []BrokerV1 `json:"brokers"`
	ClusterID      string     `json:"clusterID"`
	ControllerID   int32      `json:"controllerID"`
	Topics         []TopicV10 `json:"topics"`
}

// ApiVersions Request (Version: 0)

type ApiVersionsRequestV0 struct{}

// ApiVersions Request (Version: 3)

type ApiVersionsRequestV3 struct {
	ClientSoftwareName    string `json:"clientSoftwareName"`
	ClientSoftwareVersion string `json:"clientSoftwareVersion"`
}

// ApiVersions Response (Version: 0)

type ApiVersionsResponseApiKey struct {
	ApiKey     int16 `json:"apiKey"`
	MinVersion int16 `json:"minVersion"`
	MaxVersion int16 `json:"maxVersion"`
}

type ApiVersionsResponseV0 struct {
	ErrorCode int16                       `json:"errorCode"`
	ApiKeys   []ApiVersionsResponseApiKey `json:"apiKeys"`
}

// ApiVersions Response (Version: 1)

type ApiVersionsResponseV1 struct {
	ErrorCode      int16                       `json:"errorCode"`
	ApiKeys        []ApiVersionsResponseApiKey `json:"apiKeys"` // FIXME: `confluent-kafka-python` causes memory leak
	ThrottleTimeMs int32                       `json:"throttleTimeMs"`
}

// Produce Request (Version: 0)

// Message is a kafka message type
type MessageV0 struct {
	Codec            int8        `json:"codec"`            // codec used to compress the message contents
	CompressionLevel int         `json:"compressionLevel"` // compression level
	LogAppendTime    bool        `json:"logAppendTime"`    // the used timestamp is LogAppendTime
	Key              []byte      `json:"key"`              // the message key, may be nil
	Value            []byte      `json:"value"`            // the message contents
	Set              *MessageSet `json:"set"`              // the message set a message might wrap
	Version          int8        `json:"version"`          // v1 requires Kafka 0.10
	Timestamp        time.Time   `json:"timestamp"`        // the timestamp of the message (version 1+ only)
}

// MessageBlock represents a part of request with message
type MessageBlock struct {
	Offset int64      `json:"offset"`
	Msg    *MessageV0 `json:"msg"`
}

// MessageSet is a replacement for RecordBatch in older versions
type MessageSet struct {
	PartialTrailingMessage bool            `json:"partialTrailingMessage"` // whether the set on the wire contained an incomplete trailing MessageBlock
	OverflowMessage        bool            `json:"overflowMessage"`        // whether the set on the wire contained an overflow message
	Messages               []*MessageBlock `json:"messages"`
}

type RecordHeader struct {
	HeaderKeyLength   int8   `json:"headerKeyLength"`
	HeaderKey         string `json:"headerKey"`
	HeaderValueLength int8   `json:"headerValueLength"`
	Value             string `json:"value"`
}

// Record is kafka record type
type RecordV0 struct {
	Unknown        int8           `json:"unknown"`
	Attributes     int8           `json:"attributes"`
	TimestampDelta int8           `json:"timestampDelta"`
	OffsetDelta    int8           `json:"offsetDelta"`
	KeyLength      int8           `json:"keyLength"`
	Key            string         `json:"key"`
	ValueLen       int8           `json:"valueLen"`
	Value          string         `json:"value"`
	Headers        []RecordHeader `json:"headers"`
}

// RecordBatch are records from one kafka request
type RecordBatch struct {
	BaseOffset           int64      `json:"baseOffset"`
	BatchLength          int32      `json:"batchLength"`
	PartitionLeaderEpoch int32      `json:"partitionLeaderEpoch"`
	Magic                int8       `json:"magic"`
	Crc                  int32      `json:"crc"`
	Attributes           int16      `json:"attributes"`
	LastOffsetDelta      int32      `json:"lastOffsetDelta"`
	FirstTimestamp       int64      `json:"firstTimestamp"`
	MaxTimestamp         int64      `json:"maxTimestamp"`
	ProducerId           int64      `json:"producerId"`
	ProducerEpoch        int16      `json:"producerEpoch"`
	BaseSequence         int32      `json:"baseSequence"`
	Record               []RecordV0 `json:"record"`
}

type Records struct {
	RecordBatch RecordBatch `json:"recordBatch"`
	// TODO: Implement `MessageSet`
	// MessageSet  MessageSet
}

type PartitionData struct {
	Index   int32   `json:"index"`
	Unknown int32   `json:"unknown"`
	Records Records `json:"records"`
}

type Partitions struct {
	Length        int32         `json:"length"`
	PartitionData PartitionData `json:"partitionData"`
}

type TopicData struct {
	Topic      string     `json:"topic"`
	Partitions Partitions `json:"partitions"`
}

type ProduceRequestV0 struct {
	RequiredAcks RequiredAcks `json:"requiredAcks"`
	Timeout      int32        `json:"timeout"`
	TopicData    []TopicData  `json:"topicData"`
}

// Produce Request (Version: 3)

type ProduceRequestV3 struct {
	TransactionalID string       `json:"transactionalID"`
	RequiredAcks    RequiredAcks `json:"requiredAcks"`
	Timeout         int32        `json:"timeout"`
	TopicData       []TopicData  `json:"topicData"`
}

// Produce Response (Version: 0)

type PartitionResponseV0 struct {
	Index      int32 `json:"index"`
	ErrorCode  int16 `json:"errorCode"`
	BaseOffset int64 `json:"baseOffset"`
}

type ResponseV0 struct {
	Name               string                `json:"name"`
	PartitionResponses []PartitionResponseV0 `json:"partitionResponses"`
}

type ProduceResponseV0 struct {
	Responses []ResponseV0 `json:"responses"`
}

// Produce Response (Version: 1)

type ProduceResponseV1 struct {
	Responses      []ResponseV0 `json:"responses"`
	ThrottleTimeMs int32        `json:"throttleTimeMs"`
}

// Produce Response (Version: 2)

type PartitionResponseV2 struct {
	Index           int32 `json:"index"`
	ErrorCode       int16 `json:"errorCode"`
	BaseOffset      int64 `json:"baseOffset"`
	LogAppendTimeMs int64 `json:"logAppendTimeMs"`
}

type ResponseV2 struct {
	Name               string                `json:"name"`
	PartitionResponses []PartitionResponseV2 `json:"partitionResponses"`
}

type ProduceResponseV2 struct {
	Responses      []ResponseV2 `json:"responses"`
	ThrottleTimeMs int32        `json:"throttleTimeMs"`
}

// Produce Response (Version: 5)

type PartitionResponseV5 struct {
	Index           int32 `json:"index"`
	ErrorCode       int16 `json:"errorCode"`
	BaseOffset      int64 `json:"baseOffset"`
	LogAppendTimeMs int64 `json:"logAppendTimeMs"`
	LogStartOffset  int64 `json:"logStartOffset"`
}

type ResponseV5 struct {
	Name               string                `json:"name"`
	PartitionResponses []PartitionResponseV5 `json:"partitionResponses"`
}

type ProduceResponseV5 struct {
	Responses      []ResponseV5 `json:"responses"`
	ThrottleTimeMs int32        `json:"throttleTimeMs"`
}

// Produce Response (Version: 8)

type RecordErrors struct {
	BatchIndex             int32  `json:"batchIndex"`
	BatchIndexErrorMessage string `json:"batchIndexErrorMessage"`
}

type PartitionResponseV8 struct {
	Index           int32        `json:"index"`
	ErrorCode       int16        `json:"errorCode"`
	BaseOffset      int64        `json:"baseOffset"`
	LogAppendTimeMs int64        `json:"logAppendTimeMs"`
	LogStartOffset  int64        `json:"logStartOffset"`
	RecordErrors    RecordErrors `json:"recordErrors"`
	ErrorMessage    string       `json:"errorMessage"`
}

type ResponseV8 struct {
	Name               string                `json:"name"`
	PartitionResponses []PartitionResponseV8 `json:"partitionResponses"`
}

type ProduceResponseV8 struct {
	Responses      []ResponseV8 `json:"responses"`
	ThrottleTimeMs int32        `json:"throttleTimeMs"`
}

// Fetch Request (Version: 0)

type FetchPartitionV0 struct {
	Partition         int32 `json:"partition"`
	FetchOffset       int64 `json:"fetchOffset"`
	PartitionMaxBytes int32 `json:"partitionMaxBytes"`
}

type FetchTopicV0 struct {
	Topic      string             `json:"topic"`
	Partitions []FetchPartitionV0 `json:"partitions"`
}

type FetchRequestV0 struct {
	ReplicaId int32          `json:"replicaId"`
	MaxWaitMs int32          `json:"maxWaitMs"`
	MinBytes  int32          `json:"minBytes"`
	Topics    []FetchTopicV0 `json:"topics"`
}

// Fetch Request (Version: 3)

type FetchRequestV3 struct {
	ReplicaId int32          `json:"replicaId"`
	MaxWaitMs int32          `json:"maxWaitMs"`
	MinBytes  int32          `json:"minBytes"`
	MaxBytes  int32          `json:"maxBytes"`
	Topics    []FetchTopicV0 `json:"topics"`
}

// Fetch Request (Version: 4)

type FetchRequestV4 struct {
	ReplicaId      int32          `json:"replicaId"`
	MaxWaitMs      int32          `json:"maxWaitMs"`
	MinBytes       int32          `json:"minBytes"`
	MaxBytes       int32          `json:"maxBytes"`
	IsolationLevel int8           `json:"isolationLevel"`
	Topics         []FetchTopicV0 `json:"topics"`
}

// Fetch Request (Version: 5)

type FetchPartitionV5 struct {
	Partition         int32 `json:"partition"`
	FetchOffset       int64 `json:"fetchOffset"`
	LogStartOffset    int64 `json:"logStartOffset"`
	PartitionMaxBytes int32 `json:"partitionMaxBytes"`
}

type FetchTopicV5 struct {
	Topic      string             `json:"topic"`
	Partitions []FetchPartitionV5 `json:"partitions"`
}

type FetchRequestV5 struct {
	ReplicaId      int32          `json:"replicaId"`
	MaxWaitMs      int32          `json:"maxWaitMs"`
	MinBytes       int32          `json:"minBytes"`
	MaxBytes       int32          `json:"maxBytes"`
	IsolationLevel int8           `json:"isolationLevel"`
	Topics         []FetchTopicV5 `json:"topics"`
}

// Fetch Request (Version: 7)

type ForgottenTopicsDataV7 struct {
	Topic      string  `json:"topic"`
	Partitions []int32 `json:"partitions"`
}

type FetchRequestV7 struct {
	ReplicaId           int32                 `json:"replicaId"`
	MaxWaitMs           int32                 `json:"maxWaitMs"`
	MinBytes            int32                 `json:"minBytes"`
	MaxBytes            int32                 `json:"maxBytes"`
	IsolationLevel      int8                  `json:"isolationLevel"`
	SessionId           int32                 `json:"sessionId"`
	SessionEpoch        int32                 `json:"sessionEpoch"`
	Topics              []FetchTopicV5        `json:"topics"`
	ForgottenTopicsData ForgottenTopicsDataV7 `json:"forgottenTopicsData"`
}

// Fetch Request (Version: 9)

type FetchPartitionV9 struct {
	Partition          int32 `json:"partition"`
	CurrentLeaderEpoch int32 `json:"currentLeaderEpoch"`
	FetchOffset        int64 `json:"fetchOffset"`
	LogStartOffset     int64 `json:"logStartOffset"`
	PartitionMaxBytes  int32 `json:"partitionMaxBytes"`
}

type FetchTopicV9 struct {
	Topic      string             `json:"topic"`
	Partitions []FetchPartitionV9 `json:"partitions"`
}

type FetchRequestV9 struct {
	ReplicaId           int32                 `json:"replicaId"`
	MaxWaitMs           int32                 `json:"maxWaitMs"`
	MinBytes            int32                 `json:"minBytes"`
	MaxBytes            int32                 `json:"maxBytes"`
	IsolationLevel      int8                  `json:"isolationLevel"`
	SessionId           int32                 `json:"sessionId"`
	SessionEpoch        int32                 `json:"sessionEpoch"`
	Topics              []FetchTopicV9        `json:"topics"`
	ForgottenTopicsData ForgottenTopicsDataV7 `json:"forgottenTopicsData"`
}

// Fetch Request (Version: 11)

type FetchRequestV11 struct {
	ReplicaId           int32                 `json:"replicaId"`
	MaxWaitMs           int32                 `json:"maxWaitMs"`
	MinBytes            int32                 `json:"minBytes"`
	MaxBytes            int32                 `json:"maxBytes"`
	IsolationLevel      int8                  `json:"isolationLevel"`
	SessionId           int32                 `json:"sessionId"`
	SessionEpoch        int32                 `json:"sessionEpoch"`
	Topics              []FetchTopicV9        `json:"topics"`
	ForgottenTopicsData ForgottenTopicsDataV7 `json:"forgottenTopicsData"`
	RackId              string                `json:"rackId"`
}

// Fetch Response (Version: 0)

type PartitionResponseFetchV0 struct {
	Partition     int32   `json:"partition"`
	ErrorCode     int16   `json:"errorCode"`
	HighWatermark int64   `json:"highWatermark"`
	RecordSet     Records `json:"recordSet"`
}

type ResponseFetchV0 struct {
	Topic              string                     `json:"topic"`
	PartitionResponses []PartitionResponseFetchV0 `json:"partitionResponses"`
}

type FetchResponseV0 struct {
	Responses []ResponseFetchV0 `json:"responses"`
}

// Fetch Response (Version: 1)

type FetchResponseV1 struct {
	ThrottleTimeMs int32             `json:"throttleTimeMs"`
	Responses      []ResponseFetchV0 `json:"responses"`
}

// Fetch Response (Version: 4)

type AbortedTransactionsV4 struct {
	ProducerId  int32 `json:"producerId"`
	FirstOffset int32 `json:"firstOffset"`
}

type PartitionResponseFetchV4 struct {
	Partition           int32                 `json:"partition"`
	ErrorCode           int16                 `json:"errorCode"`
	HighWatermark       int64                 `json:"highWatermark"`
	LastStableOffset    int64                 `json:"lastStableOffset"`
	AbortedTransactions AbortedTransactionsV4 `json:"abortedTransactions"`
	RecordSet           Records               `json:"recordSet"`
}

type ResponseFetchV4 struct {
	Topic              string                     `json:"topic"`
	PartitionResponses []PartitionResponseFetchV4 `json:"partitionResponses"`
}

type FetchResponseV4 struct {
	ThrottleTimeMs int32             `json:"throttleTimeMs"`
	Responses      []ResponseFetchV4 `json:"responses"`
}

// Fetch Response (Version: 5)

type PartitionResponseFetchV5 struct {
	Partition           int32                 `json:"partition"`
	ErrorCode           int16                 `json:"errorCode"`
	HighWatermark       int64                 `json:"highWatermark"`
	LastStableOffset    int64                 `json:"lastStableOffset"`
	LogStartOffset      int64                 `json:"logStartOffset"`
	AbortedTransactions AbortedTransactionsV4 `json:"abortedTransactions"`
	RecordSet           Records               `json:"recordSet"`
}

type ResponseFetchV5 struct {
	Topic              string                     `json:"topic"`
	PartitionResponses []PartitionResponseFetchV5 `json:"partitionResponses"`
}

type FetchResponseV5 struct {
	ThrottleTimeMs int32             `json:"throttleTimeMs"`
	Responses      []ResponseFetchV5 `json:"responses"`
}

// Fetch Response (Version: 7)

type FetchResponseV7 struct {
	ThrottleTimeMs int32             `json:"throttleTimeMs"`
	ErrorCode      int16             `json:"errorCode"`
	SessionId      int32             `json:"sessionId"`
	Responses      []ResponseFetchV5 `json:"responses"`
}

// Fetch Response (Version: 11)

type PartitionResponseFetchV11 struct {
	Partition            int32                 `json:"partition"`
	ErrorCode            int16                 `json:"errorCode"`
	HighWatermark        int64                 `json:"highWatermark"`
	LastStableOffset     int64                 `json:"lastStableOffset"`
	LogStartOffset       int64                 `json:"logStartOffset"`
	AbortedTransactions  AbortedTransactionsV4 `json:"abortedTransactions"`
	PreferredReadReplica int32                 `json:"preferredReadReplica"`
	RecordSet            Records               `json:"recordSet"`
}

type ResponseFetchV11 struct {
	Topic              string                      `json:"topic"`
	PartitionResponses []PartitionResponseFetchV11 `json:"partitionResponses"`
}

type FetchResponseV11 struct {
	ThrottleTimeMs int32             `json:"throttleTimeMs"`
	ErrorCode      int16             `json:"errorCode"`
	SessionId      int32             `json:"sessionId"`
	Responses      []ResponseFetchV5 `json:"responses"`
}

// ListOffsets Request (Version: 0)

type ListOffsetsRequestPartitionV0 struct {
	PartitionIndex int32 `json:"partitionIndex"`
	Timestamp      int64 `json:"timestamp"`
	MaxNumOffsets  int32 `json:"maxNumOffsets"`
}

type ListOffsetsRequestTopicV0 struct {
	Name       string                          `json:"name"`
	Partitions []ListOffsetsRequestPartitionV0 `json:"partitions"`
}

type ListOffsetsRequestV0 struct {
	ReplicaId int32                       `json:"replicaId"`
	Topics    []ListOffsetsRequestTopicV0 `json:"topics"`
}

// ListOffsets Request (Version: 1)

type ListOffsetsRequestPartitionV1 struct {
	PartitionIndex int32 `json:"partitionIndex"`
	Timestamp      int64 `json:"timestamp"`
}

type ListOffsetsRequestTopicV1 struct {
	Name       string                          `json:"name"`
	Partitions []ListOffsetsRequestPartitionV1 `json:"partitions"`
}

type ListOffsetsRequestV1 struct {
	ReplicaId int32                       `json:"replicaId"`
	Topics    []ListOffsetsRequestTopicV1 `json:"topics"`
}

// ListOffsets Request (Version: 2)

type ListOffsetsRequestV2 struct {
	ReplicaId      int32                       `json:"replicaId"`
	IsolationLevel int8                        `json:"isolationLevel"`
	Topics         []ListOffsetsRequestTopicV1 `json:"topics"`
}

// ListOffsets Request (Version: 4)

type ListOffsetsRequestPartitionV4 struct {
	PartitionIndex     int32 `json:"partitionIndex"`
	CurrentLeaderEpoch int32 `json:"currentLeaderEpoch"`
	Timestamp          int64 `json:"timestamp"`
}

type ListOffsetsRequestTopicV4 struct {
	Name       string                          `json:"name"`
	Partitions []ListOffsetsRequestPartitionV4 `json:"partitions"`
}

type ListOffsetsRequestV4 struct {
	ReplicaId int32                       `json:"replicaId"`
	Topics    []ListOffsetsRequestTopicV4 `json:"topics"`
}

// ListOffsets Response (Version: 0)

type ListOffsetsResponsePartitionV0 struct {
	PartitionIndex  int32 `json:"partitionIndex"`
	ErrorCode       int16 `json:"errorCode"`
	OldStyleOffsets int64 `json:"oldStyleOffsets"`
}

type ListOffsetsResponseTopicV0 struct {
	Name       string                           `json:"name"`
	Partitions []ListOffsetsResponsePartitionV0 `json:"partitions"`
}

type ListOffsetsResponseV0 struct {
	Topics []ListOffsetsResponseTopicV0 `json:"topics"`
}

// ListOffsets Response (Version: 1)

type ListOffsetsResponsePartitionV1 struct {
	PartitionIndex int32 `json:"partitionIndex"`
	ErrorCode      int16 `json:"errorCode"`
	Timestamp      int64 `json:"timestamp"`
	Offset         int64 `json:"offset"`
}

type ListOffsetsResponseTopicV1 struct {
	Name       string                           `json:"name"`
	Partitions []ListOffsetsResponsePartitionV1 `json:"partitions"`
}

type ListOffsetsResponseV1 struct {
	Topics []ListOffsetsResponseTopicV1 `json:"topics"`
}

// ListOffsets Response (Version: 2)

type ListOffsetsResponseV2 struct {
	ThrottleTimeMs int32                        `json:"throttleTimeMs"`
	Topics         []ListOffsetsResponseTopicV1 `json:"topics"`
}

// ListOffsets Response (Version: 4)

type ListOffsetsResponsePartitionV4 struct {
	PartitionIndex int32 `json:"partitionIndex"`
	ErrorCode      int16 `json:"errorCode"`
	Timestamp      int64 `json:"timestamp"`
	Offset         int64 `json:"offset"`
	LeaderEpoch    int32 `json:"leaderEpoch"`
}

type ListOffsetsResponseTopicV4 struct {
	Name       string                           `json:"name"`
	Partitions []ListOffsetsResponsePartitionV4 `json:"partitions"`
}

type ListOffsetsResponseV4 struct {
	Topics []ListOffsetsResponseTopicV4 `json:"topics"`
}

// CreateTopics Request (Version: 0)

type AssignmentsV0 struct {
	PartitionIndex int32   `json:"partitionIndex"`
	BrokerIds      []int32 `json:"brokerIds"`
}

type CreateTopicsRequestConfigsV0 struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateTopicsRequestTopicV0 struct {
	Name              string                         `json:"name"`
	NumPartitions     int32                          `json:"numPartitions"`
	ReplicationFactor int16                          `json:"replicationFactor"`
	Assignments       []AssignmentsV0                `json:"assignments"`
	Configs           []CreateTopicsRequestConfigsV0 `json:"configs"`
}

type CreateTopicsRequestV0 struct {
	Topics    []CreateTopicsRequestTopicV0 `json:"topics"`
	TimeoutMs int32                        `json:"timeoutMs"`
}

// CreateTopics Request (Version: 1)

type CreateTopicsRequestV1 struct {
	Topics       []CreateTopicsRequestTopicV0 `json:"topics"`
	TimeoutMs    int32                        `json:"timeoutMs"`
	ValidateOnly bool                         `json:"validateOnly"`
}

// CreateTopics Response (Version: 0)

type CreateTopicsResponseTopicV0 struct {
	Name      string `json:"name"`
	ErrorCode int16  `json:"errorCode"`
}

type CreateTopicsResponseV0 struct {
	Topics []CreateTopicsResponseTopicV0 `json:"topics"`
}

// CreateTopics Response (Version: 1)

type CreateTopicsResponseTopicV1 struct {
	Name         string `json:"name"`
	ErrorCode    int16  `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type CreateTopicsResponseV1 struct {
	Topics []CreateTopicsResponseTopicV1 `json:"topics"`
}

// CreateTopics Response (Version: 2)

type CreateTopicsResponseV2 struct {
	ThrottleTimeMs int32                         `json:"throttleTimeMs"`
	Topics         []CreateTopicsResponseTopicV1 `json:"topics"`
}

// CreateTopics Response (Version: 5)

type CreateTopicsResponseConfigsV5 struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	ReadOnly     bool   `json:"readOnly"`
	ConfigSource int8   `json:"configSource"`
	IsSensitive  bool   `json:"isSensitive"`
}

type CreateTopicsResponseTopicV5 struct {
	Name              string                          `json:"name"`
	ErrorCode         int16                           `json:"errorCode"`
	ErrorMessage      string                          `json:"errorMessage"`
	NumPartitions     int32                           `json:"numPartitions"`
	ReplicationFactor int16                           `json:"replicationFactor"`
	Configs           []CreateTopicsResponseConfigsV5 `json:"configs"`
}

type CreateTopicsResponseV5 struct {
	ThrottleTimeMs int32                         `json:"throttleTimeMs"`
	Topics         []CreateTopicsResponseTopicV5 `json:"topics"`
}

// CreateTopics Response (Version: 7)

type CreateTopicsResponseTopicV7 struct {
	Name              string                          `json:"name"`
	TopicID           UUID                            `json:"topicID"`
	ErrorCode         int16                           `json:"errorCode"`
	ErrorMessage      string                          `json:"errorMessage"`
	NumPartitions     int32                           `json:"numPartitions"`
	ReplicationFactor int16                           `json:"replicationFactor"`
	Configs           []CreateTopicsResponseConfigsV5 `json:"configs"`
}

type CreateTopicsResponseV7 struct {
	ThrottleTimeMs int32                         `json:"throttleTimeMs"`
	Topics         []CreateTopicsResponseTopicV7 `json:"topics"`
}

// DeleteTopics Request (Version: 0)

type DeleteTopicsRequestV0 struct {
	TopicNames []string `json:"topicNames"`
	TimeoutMs  int32    `json:"timeoutMs"`
}

// DeleteTopics Request (Version: 6)

type DeleteTopicsRequestTopicV6 struct {
	Name string `json:"name"`
	UUID UUID   `json:"uuid"`
}

type DeleteTopicsRequestV6 struct {
	Topics    []DeleteTopicsRequestTopicV6 `json:"topics"`
	TimeoutMs int32                        `json:"timeoutMs"`
}

// DeleteTopics Response (Version: 0)

type DeleteTopicsReponseResponseV0 struct {
	Name      string `json:"name"`
	ErrorCode int16  `json:"errorCode"`
}

type DeleteTopicsReponseV0 struct {
	Responses []DeleteTopicsReponseResponseV0 `json:"responses"`
}

// DeleteTopics Response (Version: 1)

type DeleteTopicsReponseV1 struct {
	ThrottleTimeMs int32                           `json:"throttleTimeMs"`
	Responses      []DeleteTopicsReponseResponseV0 `json:"responses"`
}

// DeleteTopics Response (Version: 5)

type DeleteTopicsReponseResponseV5 struct {
	Name         string `json:"name"`
	ErrorCode    int16  `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type DeleteTopicsReponseV5 struct {
	ThrottleTimeMs int32                           `json:"throttleTimeMs"`
	Responses      []DeleteTopicsReponseResponseV5 `json:"responses"`
}

// DeleteTopics Response (Version: 6)

type DeleteTopicsReponseResponseV6 struct {
	Name         string `json:"name"`
	TopicID      UUID   `json:"topicID"`
	ErrorCode    int16  `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type DeleteTopicsReponseV6 struct {
	ThrottleTimeMs int32                           `json:"throttleTimeMs"`
	Responses      []DeleteTopicsReponseResponseV6 `json:"responses"`
}
