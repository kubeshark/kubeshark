package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type KafkaPayload struct {
	Type string
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
		"type":  "table",
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
		"type":  "table",
		"title": "Response Header",
		"data":  string(requestHeader),
	})

	return rep
}

func representMetadataRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["Topics"].([]interface{}))
	allowAutoTopicCreation := ""
	includeClusterAuthorizedOperations := ""
	includeTopicAuthorizedOperations := ""
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
			"value": string(topics),
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
		"type":  "table",
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
		"type":  "table",
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
		"type":  "table",
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}

func representApiVersionsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["Payload"].(map[string]interface{})
	apiKeys, _ := json.Marshal(payload["ApiKeys"].([]interface{}))
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
			"value": string(apiKeys),
		},
		{
			"name":  "Throttle Time (ms)",
			"value": throttleTimeMs,
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Payload",
		"data":  string(repPayload),
	})

	return rep
}
