package kafka

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type Request struct {
	Size          int32       `json:"size"`
	ApiKeyName    string      `json:"apiKeyName"`
	ApiKey        ApiKey      `json:"apiKey"`
	ApiVersion    int16       `json:"apiVersion"`
	CorrelationID int32       `json:"correlationID"`
	ClientID      string      `json:"clientID"`
	Payload       interface{} `json:"payload"`
	CaptureTime   time.Time   `json:"captureTime"`
}

func ReadRequest(r io.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, captureTime time.Time, reqResMatcher *requestResponseMatcher) (apiKey ApiKey, apiVersion int16, err error) {
	d := &decoder{reader: r, remain: 4}
	size := d.readInt32()

	if size > 1000000 {
		return 0, 0, fmt.Errorf("A Kafka message cannot be bigger than 1MB")
	}

	if size < 8 {
		if size == 0 {
			return 0, 0, io.EOF
		}
		return 0, 0, fmt.Errorf("A Kafka request header cannot be smaller than 8 bytes")
	}

	if err = d.err; err != nil {
		err = dontExpectEOF(err)
		return 0, 0, err
	}

	d.remain = int(size)
	apiKey = ApiKey(d.readInt16())
	apiVersion = d.readInt16()
	correlationID := d.readInt32()
	clientID := d.readString()

	if i := int(apiKey); i < 0 || i >= numApis {
		err = fmt.Errorf("unsupported api key: %d", i)
		return apiKey, apiVersion, err
	}

	if err = d.err; err != nil {
		err = dontExpectEOF(err)
		return apiKey, apiVersion, err
	}

	var payload interface{}

	switch apiKey {
	case Metadata:
		var mt interface{}
		var metadataRequest interface{}
		if apiVersion >= v11 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV11{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV11{}
		} else if apiVersion >= v10 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV10{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV10{}
		} else if apiVersion >= v8 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV8{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV8{}
		} else if apiVersion >= v4 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV4{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV4{}
		} else {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV0{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(metadataRequest))
		payload = metadataRequest
	case ApiVersions:
		var mt interface{}
		var apiVersionsRequest interface{}
		if apiVersion >= v3 {
			types := makeTypes(reflect.TypeOf(&ApiVersionsRequestV3{}).Elem())
			mt = types[0]
			apiVersionsRequest = &ApiVersionsRequestV3{}
		} else {
			types := makeTypes(reflect.TypeOf(&ApiVersionsRequestV0{}).Elem())
			mt = types[0]
			apiVersionsRequest = &ApiVersionsRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(apiVersionsRequest))
		payload = apiVersionsRequest
	case Produce:
		var mt interface{}
		var produceRequest interface{}
		if apiVersion >= v3 {
			types := makeTypes(reflect.TypeOf(&ProduceRequestV3{}).Elem())
			mt = types[0]
			produceRequest = &ProduceRequestV3{}
		} else {
			types := makeTypes(reflect.TypeOf(&ProduceRequestV0{}).Elem())
			mt = types[0]
			produceRequest = &ProduceRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(produceRequest))
		payload = produceRequest
	case Fetch:
		var mt interface{}
		var fetchRequest interface{}
		if apiVersion >= 11 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV11{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV11{}
		} else if apiVersion >= v9 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV9{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV9{}
		} else if apiVersion >= v7 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV7{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV7{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV5{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV5{}
		} else if apiVersion >= v4 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV4{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV4{}
		} else if apiVersion >= v3 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV3{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV3{}
		} else {
			types := makeTypes(reflect.TypeOf(&FetchRequestV0{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(fetchRequest))
		payload = fetchRequest
	case ListOffsets:
		var mt interface{}
		var listOffsetsRequest interface{}
		if apiVersion >= v4 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV4{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV4{}
		} else if apiVersion >= v2 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV2{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV2{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV1{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV0{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(listOffsetsRequest))
		payload = listOffsetsRequest
	case CreateTopics:
		var mt interface{}
		var createTopicsRequest interface{}
		if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&CreateTopicsRequestV1{}).Elem())
			mt = types[0]
			createTopicsRequest = &CreateTopicsRequestV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&CreateTopicsRequestV0{}).Elem())
			mt = types[0]
			createTopicsRequest = &CreateTopicsRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(createTopicsRequest))
		payload = createTopicsRequest
	case DeleteTopics:
		var mt interface{}
		var deleteTopicsRequest interface{}
		if apiVersion >= v6 {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsRequestV6{}).Elem())
			mt = types[0]
			deleteTopicsRequest = &DeleteTopicsRequestV6{}
		} else {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsRequestV0{}).Elem())
			mt = types[0]
			deleteTopicsRequest = &DeleteTopicsRequestV0{}
		}
		mt.(messageType).decode(d, valueOf(deleteTopicsRequest))
		payload = deleteTopicsRequest
	default:
		return apiKey, 0, fmt.Errorf("(Request) Not implemented: %s", apiKey)
	}

	request := &Request{
		Size:          size,
		ApiKeyName:    apiNames[apiKey],
		ApiKey:        apiKey,
		ApiVersion:    apiVersion,
		CorrelationID: correlationID,
		ClientID:      clientID,
		CaptureTime:   captureTime,
		Payload:       payload,
	}

	key := fmt.Sprintf(
		"%s_%s_%s_%s_%d",
		tcpID.SrcIP,
		tcpID.SrcPort,
		tcpID.DstIP,
		tcpID.DstPort,
		correlationID,
	)
	reqResMatcher.registerRequest(key, request)

	d.discardAll()

	return apiKey, apiVersion, nil
}
