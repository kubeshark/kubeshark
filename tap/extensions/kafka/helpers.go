package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/up9inc/mizu/tap/api"
)

type KafkaPayload struct {
	Data interface{}
}

type KafkaPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h KafkaPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Data)
}

type KafkaWrapper struct {
	Method  string      `json:"method"`
	Url     string      `json:"url"`
	Details interface{} `json:"details"`
}

func representRequestHeader(data map[string]interface{}, rep []interface{}) []interface{} {
	requestHeader, _ := json.Marshal([]map[string]string{
		{
			"name":  "ApiKey",
			"value": apiNames[int(data["ApiKey"].(float64))],
		},
		{
			"name":  "ApiVersion",
			"value": fmt.Sprintf("%d", int(data["ApiVersion"].(float64))),
		},
		{
			"name":  "Client ID",
			"value": data["ClientID"].(string),
		},
		{
			"name":  "Correlation ID",
			"value": fmt.Sprintf("%d", int(data["CorrelationID"].(float64))),
		},
		{
			"name":  "Size",
			"value": fmt.Sprintf("%d", int(data["Size"].(float64))),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Request Header",
		"data":  string(requestHeader),
	})

	return rep
}

func representResponseHeader(data map[string]interface{}, rep []interface{}) []interface{} {
	requestHeader, _ := json.Marshal([]map[string]string{
		{
			"name":  "Correlation ID",
			"value": fmt.Sprintf("%d", int(data["CorrelationID"].(float64))),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Response Header",
		"data":  string(requestHeader),
	})

	return rep
}

func representMetadataRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics := ""
	allowAutoTopicCreation := ""
	includeClusterAuthorizedOperations := ""
	includeTopicAuthorizedOperations := ""
	if payload["Topics"] != nil {
		x, _ := json.Marshal(payload["Topics"].([]interface{}))
		topics = string(x)
	}
	if payload["AllowAutoTopicCreation"] != nil {
		allowAutoTopicCreation = strconv.FormatBool(payload["AllowAutoTopicCreation"].(bool))
	}
	if payload["IncludeClusterAuthorizedOperations"] != nil {
		includeClusterAuthorizedOperations = strconv.FormatBool(payload["IncludeClusterAuthorizedOperations"].(bool))
	}
	if payload["IncludeTopicAuthorizedOperations"] != nil {
		includeTopicAuthorizedOperations = strconv.FormatBool(payload["IncludeTopicAuthorizedOperations"].(bool))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Topics",
			"value": topics,
		},
		{
			"name":  "Allow Auto Topic Creation",
			"value": allowAutoTopicCreation,
		},
		{
			"name":  "Include Cluster Authorized Operations",
			"value": includeClusterAuthorizedOperations,
		},
		{
			"name":  "Include Topic Authorized Operations",
			"value": includeTopicAuthorizedOperations,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representMetadataResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	brokers, _ := json.Marshal(payload["Brokers"].([]interface{}))
	controllerID := ""
	clusterID := ""
	throttleTimeMs := ""
	clusterAuthorizedOperations := ""
	if payload["ControllerID"] != nil {
		controllerID = fmt.Sprintf("%d", int(payload["ControllerID"].(float64)))
	}
	if payload["ClusterID"] != nil {
		clusterID = payload["ClusterID"].(string)
	}
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	if payload["ClusterAuthorizedOperations"] != nil {
		clusterAuthorizedOperations = fmt.Sprintf("%d", int(payload["ClusterAuthorizedOperations"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
		{
			"name":  "Brokers",
			"value": string(brokers),
		},
		{
			"name":  "Cluster ID",
			"value": clusterID,
		},
		{
			"name":  "Controller ID",
			"value": controllerID,
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
		{
			"name":  "Cluster Authorized Operations",
			"value": clusterAuthorizedOperations,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representApiVersionsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	clientSoftwareName := ""
	clientSoftwareVersion := ""
	if payload["ClientSoftwareName"] != nil {
		clientSoftwareName = payload["ClientSoftwareName"].(string)
	}
	if payload["ClientSoftwareVersion"] != nil {
		clientSoftwareVersion = payload["ClientSoftwareVersion"].(string)
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Client Software Name",
			"value": clientSoftwareName,
		},
		{
			"name":  "Client Software Version",
			"value": clientSoftwareVersion,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representApiVersionsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	apiKeys := ""
	if payload["TopicNames"] != nil {
		x, _ := json.Marshal(payload["ApiKeys"].([]interface{}))
		apiKeys = string(x)
	}
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Error Code",
			"value": fmt.Sprintf("%d", int(payload["ErrorCode"].(float64))),
		},
		{
			"name":  "ApiKeys",
			"value": apiKeys,
		},
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representProduceRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topicData := ""
	_topicData := payload["TopicData"]
	if _topicData != nil {
		x, _ := json.Marshal(_topicData.([]interface{}))
		topicData = string(x)
	}
	transactionalID := ""
	if payload["TransactionalID"] != nil {
		transactionalID = payload["TransactionalID"].(string)
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Transactional ID",
			"value": transactionalID,
		},
		{
			"name":  "Required Acknowledgements",
			"value": fmt.Sprintf("%d", int(payload["RequiredAcks"].(float64))),
		},
		{
			"name":  "Timeout",
			"value": fmt.Sprintf("%d", int(payload["Timeout"].(float64))),
		},
		{
			"name":  "Topic Data",
			"value": topicData,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representProduceResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	responses, _ := json.Marshal(payload["Responses"].([]interface{}))
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Responses",
			"value": string(responses),
		},
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representFetchRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	replicaId := ""
	if payload["ReplicaId"] != nil {
		replicaId = fmt.Sprintf("%d", int(payload["ReplicaId"].(float64)))
	}
	maxBytes := ""
	if payload["MaxBytes"] != nil {
		maxBytes = fmt.Sprintf("%d", int(payload["MaxBytes"].(float64)))
	}
	isolationLevel := ""
	if payload["IsolationLevel"] != nil {
		isolationLevel = fmt.Sprintf("%d", int(payload["IsolationLevel"].(float64)))
	}
	sessionId := ""
	if payload["SessionId"] != nil {
		sessionId = fmt.Sprintf("%d", int(payload["SessionId"].(float64)))
	}
	sessionEpoch := ""
	if payload["SessionEpoch"] != nil {
		sessionEpoch = fmt.Sprintf("%d", int(payload["SessionEpoch"].(float64)))
	}
	forgottenTopicsData := ""
	if payload["ForgottenTopicsData"] != nil {
		x, _ := json.Marshal(payload["ForgottenTopicsData"].(map[string]interface{}))
		forgottenTopicsData = string(x)
	}
	rackId := ""
	if payload["RackId"] != nil {
		rackId = fmt.Sprintf("%d", int(payload["RackId"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Replica ID",
			"value": replicaId,
		},
		{
			"name":  "Maximum Wait (ms)",
			"value": fmt.Sprintf("%d", int(payload["MaxWaitMs"].(float64))),
		},
		{
			"name":  "Minimum Bytes",
			"value": fmt.Sprintf("%d", int(payload["MinBytes"].(float64))),
		},
		{
			"name":  "Maximum Bytes",
			"value": maxBytes,
		},
		{
			"name":  "Isolation Level",
			"value": isolationLevel,
		},
		{
			"name":  "Session ID",
			"value": sessionId,
		},
		{
			"name":  "Session Epoch",
			"value": sessionEpoch,
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
		{
			"name":  "Forgotten Topics Data",
			"value": forgottenTopicsData,
		},
		{
			"name":  "Rack ID",
			"value": rackId,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representFetchResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	responses, _ := json.Marshal(payload["Responses"].([]interface{}))
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	errorCode := ""
	if payload["ErrorCode"] != nil {
		errorCode = fmt.Sprintf("%d", int(payload["ErrorCode"].(float64)))
	}
	sessionId := ""
	if payload["SessionId"] != nil {
		sessionId = fmt.Sprintf("%d", int(payload["SessionId"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
		{
			"name":  "Error Code",
			"value": errorCode,
		},
		{
			"name":  "Session ID",
			"value": sessionId,
		},
		{
			"name":  "Responses",
			"value": string(responses),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representListOffsetsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Replica ID",
			"value": fmt.Sprintf("%d", int(payload["ReplicaId"].(float64))),
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representListOffsetsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representCreateTopicsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	validateOnly := ""
	if payload["ValidateOnly"] != nil {
		validateOnly = strconv.FormatBool(payload["ValidateOnly"].(bool))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Topics",
			"value": string(topics),
		},
		{
			"name":  "Timeout (ms)",
			"value": fmt.Sprintf("%d", int(payload["TimeoutMs"].(float64))),
		},
		{
			"name":  "Validate Only",
			"value": validateOnly,
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representCreateTopicsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representDeleteTopicsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics := ""
	if payload["Topics"] != nil {
		x, _ := json.Marshal(payload["Topics"].([]interface{}))
		topics = string(x)
	}
	topicNames := ""
	if payload["TopicNames"] != nil {
		x, _ := json.Marshal(payload["TopicNames"].([]interface{}))
		topicNames = string(x)
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "TopicNames",
			"value": string(topicNames),
		},
		{
			"name":  "Topics",
			"value": string(topics),
		},
		{
			"name":  "Timeout (ms)",
			"value": fmt.Sprintf("%d", int(payload["TimeoutMs"].(float64))),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representDeleteTopicsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	responses, _ := json.Marshal(payload["Responses"].([]interface{}))
	throttleTimeMs := ""
	if payload["ThrottleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["ThrottleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]map[string]string{
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
		{
			"name":  "Responses",
			"value": string(responses),
		},
	})
	rep = append(rep, map[string]string{
		"type":  api.TABLE,
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}
