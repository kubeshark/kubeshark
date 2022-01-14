package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
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
	requestHeader, _ := json.Marshal([]api.TableData{
		{
			Name:     "ApiKey",
			Value:    apiNames[int(data["apiKey"].(float64))],
			Selector: `request.apiKey`,
		},
		{
			Name:     "ApiVersion",
			Value:    fmt.Sprintf("%d", int(data["apiVersion"].(float64))),
			Selector: `request.apiVersion`,
		},
		{
			Name:     "Client ID",
			Value:    data["clientID"].(string),
			Selector: `request.clientID`,
		},
		{
			Name:     "Correlation ID",
			Value:    fmt.Sprintf("%d", int(data["correlationID"].(float64))),
			Selector: `request.correlationID`,
		},
		{
			Name:     "Size",
			Value:    fmt.Sprintf("%d", int(data["size"].(float64))),
			Selector: `request.size`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Request Header",
		Data:  string(requestHeader),
	})

	return rep
}

func representResponseHeader(data map[string]interface{}, rep []interface{}) []interface{} {
	requestHeader, _ := json.Marshal([]api.TableData{
		{
			Name:     "Correlation ID",
			Value:    fmt.Sprintf("%d", int(data["correlationID"].(float64))),
			Selector: `response.correlationID`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Response Header",
		Data:  string(requestHeader),
	})

	return rep
}

func representMetadataRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics := ""
	allowAutoTopicCreation := ""
	includeClusterAuthorizedOperations := ""
	includeTopicAuthorizedOperations := ""
	if payload["topics"] != nil {
		x, _ := json.Marshal(payload["topics"].([]interface{}))
		topics = string(x)
	}
	if payload["allowAutoTopicCreation"] != nil {
		allowAutoTopicCreation = strconv.FormatBool(payload["allowAutoTopicCreation"].(bool))
	}
	if payload["includeClusterAuthorizedOperations"] != nil {
		includeClusterAuthorizedOperations = strconv.FormatBool(payload["includeClusterAuthorizedOperations"].(bool))
	}
	if payload["includeTopicAuthorizedOperations"] != nil {
		includeTopicAuthorizedOperations = strconv.FormatBool(payload["includeTopicAuthorizedOperations"].(bool))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Topics",
			Value:    topics,
			Selector: `request.payload.topics`,
		},
		{
			Name:     "Allow Auto Topic Creation",
			Value:    allowAutoTopicCreation,
			Selector: `request.payload.allowAutoTopicCreation`,
		},
		{
			Name:     "Include Cluster Authorized Operations",
			Value:    includeClusterAuthorizedOperations,
			Selector: `request.payload.includeClusterAuthorizedOperations`,
		},
		{
			Name:     "Include Topic Authorized Operations",
			Value:    includeTopicAuthorizedOperations,
			Selector: `request.payload.includeTopicAuthorizedOperations`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representMetadataResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics := ""
	if payload["topics"] != nil {
		_topics, _ := json.Marshal(payload["topics"].([]interface{}))
		topics = string(_topics)
	}
	brokers := ""
	if payload["brokers"] != nil {
		_brokers, _ := json.Marshal(payload["brokers"].([]interface{}))
		brokers = string(_brokers)
	}
	controllerID := ""
	clusterID := ""
	throttleTimeMs := ""
	clusterAuthorizedOperations := ""
	if payload["controllerID"] != nil {
		controllerID = fmt.Sprintf("%d", int(payload["controllerID"].(float64)))
	}
	if payload["clusterID"] != nil {
		clusterID = payload["clusterID"].(string)
	}
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	if payload["clusterAuthorizedOperations"] != nil {
		clusterAuthorizedOperations = fmt.Sprintf("%d", int(payload["clusterAuthorizedOperations"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
		{
			Name:     "Brokers",
			Value:    brokers,
			Selector: `response.payload.brokers`,
		},
		{
			Name:     "Cluster ID",
			Value:    clusterID,
			Selector: `response.payload.clusterID`,
		},
		{
			Name:     "Controller ID",
			Value:    controllerID,
			Selector: `response.payload.controllerID`,
		},
		{
			Name:     "Topics",
			Value:    topics,
			Selector: `response.payload.topics`,
		},
		{
			Name:     "Cluster Authorized Operations",
			Value:    clusterAuthorizedOperations,
			Selector: `response.payload.clusterAuthorizedOperations`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representApiVersionsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	clientSoftwareName := ""
	clientSoftwareVersion := ""
	if payload["clientSoftwareName"] != nil {
		clientSoftwareName = payload["clientSoftwareName"].(string)
	}
	if payload["clientSoftwareVersion"] != nil {
		clientSoftwareVersion = payload["clientSoftwareVersion"].(string)
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Client Software Name",
			Value:    clientSoftwareName,
			Selector: `request.payload.clientSoftwareName`,
		},
		{
			Name:     "Client Software Version",
			Value:    clientSoftwareVersion,
			Selector: `request.payload.clientSoftwareVersion`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representApiVersionsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	apiKeys := ""
	if payload["apiKeys"] != nil {
		x, _ := json.Marshal(payload["apiKeys"].([]interface{}))
		apiKeys = string(x)
	}
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Error Code",
			Value:    fmt.Sprintf("%d", int(payload["errorCode"].(float64))),
			Selector: `response.payload.errorCode`,
		},
		{
			Name:     "ApiKeys",
			Value:    apiKeys,
			Selector: `response.payload.apiKeys`,
		},
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representProduceRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topicData := payload["topicData"]
	transactionalID := ""
	if payload["transactionalID"] != nil {
		transactionalID = payload["transactionalID"].(string)
	}
	repTransactionDetails, _ := json.Marshal([]api.TableData{
		{
			Name:     "Transactional ID",
			Value:    transactionalID,
			Selector: `request.payload.transactionalID`,
		},
		{
			Name:     "Required Acknowledgements",
			Value:    fmt.Sprintf("%d", int(payload["requiredAcks"].(float64))),
			Selector: `request.payload.requiredAcks`,
		},
		{
			Name:     "Timeout",
			Value:    fmt.Sprintf("%d", int(payload["timeout"].(float64))),
			Selector: `request.payload.timeout`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Transaction Details",
		Data:  string(repTransactionDetails),
	})

	if topicData != nil {
		for _, _topic := range topicData.([]interface{}) {
			topic := _topic.(map[string]interface{})
			topicName := topic["topic"].(string)
			partitions := topic["partitions"].(map[string]interface{})
			partitionsJson, err := json.Marshal(partitions)
			if err != nil {
				return rep
			}

			repPartitions, _ := json.Marshal([]api.TableData{
				{
					Name:     "Length",
					Value:    partitions["length"],
					Selector: `request.payload.transactionalID`,
				},
			})
			rep = append(rep, api.SectionData{
				Type:  api.TABLE,
				Title: fmt.Sprintf("Partitions (topic: %s)", topicName),
				Data:  string(repPartitions),
			})

			obj, err := oj.ParseString(string(partitionsJson))
			recordBatchPath, err := jp.ParseString(`partitionData.records.recordBatch`)
			recordBatchresults := recordBatchPath.Get(obj)
			if len(recordBatchresults) > 0 {
				rep = append(rep, api.SectionData{
					Type:  api.TABLE,
					Title: fmt.Sprintf("Record Batch (topic: %s)", topicName),
					Data:  representMapAsTable(recordBatchresults[0].(map[string]interface{}), `request.payload.topicData.partitions.partitionData.records.recordBatch`),
				})
			}

			recordsPath, err := jp.ParseString(`partitionData.records.recordBatch.record`)
			recordsResults := recordsPath.Get(obj)
			if len(recordsResults) > 0 {
				records := recordsResults[0].([]interface{})
				for i, _record := range records {
					record := _record.(map[string]interface{})
					value := record["value"]
					delete(record, "value")

					rep = append(rep, api.SectionData{
						Type:  api.TABLE,
						Title: fmt.Sprintf("Record [%d] Details (topic: %s)", i, topicName),
						Data:  representMapAsTable(recordBatchresults[0].(map[string]interface{}), fmt.Sprintf(`request.payload.topicData.partitions.partitionData.records.recordBatch.record[%d]`, i)),
					})

					rep = append(rep, api.SectionData{
						Type:     api.BODY,
						Title:    fmt.Sprintf("Record [%d] Value", i),
						Data:     value.(string),
						Selector: fmt.Sprintf(`request.payload.topicData.partitions.partitionData.records.recordBatch.record[%d].value`, i),
					})
				}
			}
		}
	}

	return rep
}

func representProduceResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	responses := ""
	if payload["responses"] != nil {
		_responses, _ := json.Marshal(payload["responses"].([]interface{}))
		responses = string(_responses)
	}
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Responses",
			Value:    string(responses),
			Selector: `response.payload.responses`,
		},
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representFetchRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics := ""
	if payload["topics"] != nil {
		_topics, _ := json.Marshal(payload["topics"].([]interface{}))
		topics = string(_topics)
	}
	replicaId := ""
	if payload["replicaId"] != nil {
		replicaId = fmt.Sprintf("%d", int(payload["replicaId"].(float64)))
	}
	maxBytes := ""
	if payload["maxBytes"] != nil {
		maxBytes = fmt.Sprintf("%d", int(payload["maxBytes"].(float64)))
	}
	isolationLevel := ""
	if payload["isolationLevel"] != nil {
		isolationLevel = fmt.Sprintf("%d", int(payload["isolationLevel"].(float64)))
	}
	sessionId := ""
	if payload["sessionId"] != nil {
		sessionId = fmt.Sprintf("%d", int(payload["sessionId"].(float64)))
	}
	sessionEpoch := ""
	if payload["sessionEpoch"] != nil {
		sessionEpoch = fmt.Sprintf("%d", int(payload["sessionEpoch"].(float64)))
	}
	forgottenTopicsData := ""
	if payload["forgottenTopicsData"] != nil {
		x, _ := json.Marshal(payload["forgottenTopicsData"].(map[string]interface{}))
		forgottenTopicsData = string(x)
	}
	rackId := ""
	if payload["rackId"] != nil {
		rackId = payload["rackId"].(string)
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Replica ID",
			Value:    replicaId,
			Selector: `request.payload.replicaId`,
		},
		{
			Name:     "Maximum Wait (ms)",
			Value:    fmt.Sprintf("%d", int(payload["maxWaitMs"].(float64))),
			Selector: `request.payload.maxWaitMs`,
		},
		{
			Name:     "Minimum Bytes",
			Value:    fmt.Sprintf("%d", int(payload["minBytes"].(float64))),
			Selector: `request.payload.minBytes`,
		},
		{
			Name:     "Maximum Bytes",
			Value:    maxBytes,
			Selector: `request.payload.maxBytes`,
		},
		{
			Name:     "Isolation Level",
			Value:    isolationLevel,
			Selector: `request.payload.isolationLevel`,
		},
		{
			Name:     "Session ID",
			Value:    sessionId,
			Selector: `request.payload.sessionId`,
		},
		{
			Name:     "Session Epoch",
			Value:    sessionEpoch,
			Selector: `request.payload.sessionEpoch`,
		},
		{
			Name:     "Topics",
			Value:    topics,
			Selector: `request.payload.topics`,
		},
		{
			Name:     "Forgotten Topics Data",
			Value:    forgottenTopicsData,
			Selector: `request.payload.forgottenTopicsData`,
		},
		{
			Name:     "Rack ID",
			Value:    rackId,
			Selector: `request.payload.rackId`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representFetchResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	responses := ""
	if payload["responses"] != nil {
		_responses, _ := json.Marshal(payload["responses"].([]interface{}))
		responses = string(_responses)
	}
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	errorCode := ""
	if payload["errorCode"] != nil {
		errorCode = fmt.Sprintf("%d", int(payload["errorCode"].(float64)))
	}
	sessionId := ""
	if payload["sessionId"] != nil {
		sessionId = fmt.Sprintf("%d", int(payload["sessionId"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
		{
			Name:     "Error Code",
			Value:    errorCode,
			Selector: `response.payload.errorCode`,
		},
		{
			Name:     "Session ID",
			Value:    sessionId,
			Selector: `response.payload.sessionId`,
		},
		{
			Name:     "Responses",
			Value:    responses,
			Selector: `response.payload.responses`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representListOffsetsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics := ""
	if payload["topics"] != nil {
		_topics, _ := json.Marshal(payload["topics"].([]interface{}))
		topics = string(_topics)
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Replica ID",
			Value:    fmt.Sprintf("%d", int(payload["replicaId"].(float64))),
			Selector: `request.payload.replicaId`,
		},
		{
			Name:     "Topics",
			Value:    topics,
			Selector: `request.payload.topics`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representListOffsetsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["topics"].([]interface{}))
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
		{
			Name:     "Topics",
			Value:    string(topics),
			Selector: `response.payload.topics`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representCreateTopicsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["topics"].([]interface{}))
	validateOnly := ""
	if payload["validateOnly"] != nil {
		validateOnly = strconv.FormatBool(payload["validateOnly"].(bool))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Topics",
			Value:    string(topics),
			Selector: `request.payload.topics`,
		},
		{
			Name:     "Timeout (ms)",
			Value:    fmt.Sprintf("%d", int(payload["timeoutMs"].(float64))),
			Selector: `request.payload.timeoutMs`,
		},
		{
			Name:     "Validate Only",
			Value:    validateOnly,
			Selector: `request.payload.validateOnly`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representCreateTopicsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics, _ := json.Marshal(payload["topics"].([]interface{}))
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
		{
			Name:     "Topics",
			Value:    string(topics),
			Selector: `response.payload.topics`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representDeleteTopicsRequest(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representRequestHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	topics := ""
	if payload["topics"] != nil {
		x, _ := json.Marshal(payload["topics"].([]interface{}))
		topics = string(x)
	}
	topicNames := ""
	if payload["topicNames"] != nil {
		x, _ := json.Marshal(payload["topicNames"].([]interface{}))
		topicNames = string(x)
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "TopicNames",
			Value:    string(topicNames),
			Selector: `request.payload.topicNames`,
		},
		{
			Name:     "Topics",
			Value:    string(topics),
			Selector: `request.payload.topics`,
		},
		{
			Name:     "Timeout (ms)",
			Value:    fmt.Sprintf("%d", int(payload["timeoutMs"].(float64))),
			Selector: `request.payload.timeoutMs`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representDeleteTopicsResponse(data map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	rep = representResponseHeader(data, rep)

	payload := data["payload"].(map[string]interface{})
	responses, _ := json.Marshal(payload["responses"].([]interface{}))
	throttleTimeMs := ""
	if payload["throttleTimeMs"] != nil {
		throttleTimeMs = fmt.Sprintf("%d", int(payload["throttleTimeMs"].(float64)))
	}
	repPayload, _ := json.Marshal([]api.TableData{
		{
			Name:     "Throttle Time (ms)",
			Value:    throttleTimeMs,
			Selector: `response.payload.throttleTimeMs`,
		},
		{
			Name:     "Responses",
			Value:    string(responses),
			Selector: `response.payload.responses`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Payload",
		Data:  string(repPayload),
	})

	return rep
}

func representMapAsTable(mapData map[string]interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for key, value := range mapData {
		switch reflect.ValueOf(value).Kind() {
		case reflect.Map:
			fallthrough
		case reflect.Slice:
			continue
		default:
			selector := fmt.Sprintf("%s[\"%s\"]", selectorPrefix, key)
			table = append(table, api.TableData{
				Name:     strings.Join(camelcase.Split(strings.Title(key)), " "),
				Value:    value,
				Selector: selector,
			})
		}
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}
