package kafka

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type Response struct {
	Size          int32       `json:"size"`
	CorrelationID int32       `json:"correlationID"`
	Payload       interface{} `json:"payload"`
	CaptureTime   time.Time   `json:"captureTime"`
}

func ReadResponse(r io.Reader, capture api.Capture, tcpID *api.TcpID, counterPair *api.CounterPair, captureTime time.Time, emitter api.Emitter, reqResMatcher *requestResponseMatcher) (err error) {
	d := &decoder{reader: r, remain: 4}
	size := d.readInt32()

	if size > 1000000 {
		return fmt.Errorf("A Kafka message cannot be bigger than 1MB")
	}

	if size < 4 {
		if size == 0 {
			return io.EOF
		}
		return fmt.Errorf("A Kafka response header cannot be smaller than 8 bytes")
	}

	if err = d.err; err != nil {
		err = dontExpectEOF(err)
		return err
	}

	d.remain = int(size)
	correlationID := d.readInt32()
	var payload interface{}
	response := &Response{
		Size:          size,
		CorrelationID: correlationID,
		Payload:       payload,
		CaptureTime:   captureTime,
	}

	key := fmt.Sprintf(
		"%s_%s_%s_%s_%d",
		tcpID.DstIP,
		tcpID.DstPort,
		tcpID.SrcIP,
		tcpID.SrcPort,
		correlationID,
	)
	reqResPair := reqResMatcher.registerResponse(key, response)
	if reqResPair == nil {
		return fmt.Errorf("Couldn't match a Kafka response to a Kafka request in %d milliseconds!", reqResMatcher.maxTry)
	}
	apiKey := reqResPair.Request.ApiKey
	apiVersion := reqResPair.Request.ApiVersion

	switch apiKey {
	case Metadata:
		var mt interface{}
		var metadataResponse interface{}
		if apiVersion >= v11 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV11{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV11{}
		} else if apiVersion >= v10 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV10{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV10{}
		} else if apiVersion >= v8 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV8{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV8{}
		} else if apiVersion >= v7 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV7{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV7{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV5{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV5{}
		} else if apiVersion >= v3 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV3{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV3{}
		} else if apiVersion >= v2 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV2{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV2{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV1{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&MetadataResponseV0{}).Elem())
			mt = types[0]
			metadataResponse = &MetadataResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(metadataResponse))
		reqResPair.Response.Payload = metadataResponse
	case ApiVersions:
		var mt interface{}
		var apiVersionsResponse interface{}
		if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&ApiVersionsResponseV1{}).Elem())
			mt = types[0]
			apiVersionsResponse = &ApiVersionsResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&ApiVersionsResponseV0{}).Elem())
			mt = types[0]
			apiVersionsResponse = &ApiVersionsResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(apiVersionsResponse))
		reqResPair.Response.Payload = apiVersionsResponse
	case Produce:
		var mt interface{}
		var produceResponse interface{}
		if apiVersion >= v8 {
			types := makeTypes(reflect.TypeOf(&ProduceResponseV8{}).Elem())
			mt = types[0]
			produceResponse = &ProduceResponseV8{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&ProduceResponseV5{}).Elem())
			mt = types[0]
			produceResponse = &ProduceResponseV5{}
		} else if apiVersion >= v2 {
			types := makeTypes(reflect.TypeOf(&ProduceResponseV2{}).Elem())
			mt = types[0]
			produceResponse = &ProduceResponseV2{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&ProduceResponseV1{}).Elem())
			mt = types[0]
			produceResponse = &ProduceResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&ProduceResponseV0{}).Elem())
			mt = types[0]
			produceResponse = &ProduceResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(produceResponse))
		reqResPair.Response.Payload = produceResponse
	case Fetch:
		var mt interface{}
		var fetchResponse interface{}
		if apiVersion >= v11 {
			types := makeTypes(reflect.TypeOf(&FetchResponseV11{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV11{}
		} else if apiVersion >= v7 {
			types := makeTypes(reflect.TypeOf(&FetchResponseV7{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV7{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&FetchResponseV5{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV5{}
		} else if apiVersion >= v4 {
			types := makeTypes(reflect.TypeOf(&FetchResponseV4{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV4{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&FetchResponseV1{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&FetchResponseV0{}).Elem())
			mt = types[0]
			fetchResponse = &FetchResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(fetchResponse))
		reqResPair.Response.Payload = fetchResponse
	case ListOffsets:
		var mt interface{}
		var listOffsetsResponse interface{}
		if apiVersion >= v4 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsResponseV4{}).Elem())
			mt = types[0]
			listOffsetsResponse = &ListOffsetsResponseV4{}
		} else if apiVersion >= v2 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsResponseV2{}).Elem())
			mt = types[0]
			listOffsetsResponse = &ListOffsetsResponseV2{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsResponseV1{}).Elem())
			mt = types[0]
			listOffsetsResponse = &ListOffsetsResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&ListOffsetsResponseV0{}).Elem())
			mt = types[0]
			listOffsetsResponse = &ListOffsetsResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(listOffsetsResponse))
		reqResPair.Response.Payload = listOffsetsResponse
	case CreateTopics:
		var mt interface{}
		var createTopicsResponse interface{}
		if apiVersion >= v7 {
			types := makeTypes(reflect.TypeOf(&CreateTopicsResponseV0{}).Elem())
			mt = types[0]
			createTopicsResponse = &CreateTopicsResponseV0{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&CreateTopicsResponseV5{}).Elem())
			mt = types[0]
			createTopicsResponse = &CreateTopicsResponseV5{}
		} else if apiVersion >= v2 {
			types := makeTypes(reflect.TypeOf(&CreateTopicsResponseV2{}).Elem())
			mt = types[0]
			createTopicsResponse = &CreateTopicsResponseV2{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&CreateTopicsResponseV1{}).Elem())
			mt = types[0]
			createTopicsResponse = &CreateTopicsResponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&CreateTopicsResponseV0{}).Elem())
			mt = types[0]
			createTopicsResponse = &CreateTopicsResponseV0{}
		}
		mt.(messageType).decode(d, valueOf(createTopicsResponse))
		reqResPair.Response.Payload = createTopicsResponse
	case DeleteTopics:
		var mt interface{}
		var deleteTopicsResponse interface{}
		if apiVersion >= v6 {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsReponseV6{}).Elem())
			mt = types[0]
			deleteTopicsResponse = &DeleteTopicsReponseV6{}
		} else if apiVersion >= v5 {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsReponseV5{}).Elem())
			mt = types[0]
			deleteTopicsResponse = &DeleteTopicsReponseV5{}
		} else if apiVersion >= v1 {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsReponseV1{}).Elem())
			mt = types[0]
			deleteTopicsResponse = &DeleteTopicsReponseV1{}
		} else {
			types := makeTypes(reflect.TypeOf(&DeleteTopicsReponseV0{}).Elem())
			mt = types[0]
			deleteTopicsResponse = &DeleteTopicsReponseV0{}
		}
		mt.(messageType).decode(d, valueOf(deleteTopicsResponse))
		reqResPair.Response.Payload = deleteTopicsResponse
	default:
		return fmt.Errorf("(Response) Not implemented: %s", apiKey)
	}

	connectionInfo := &api.ConnectionInfo{
		ClientIP:   tcpID.DstIP,
		ClientPort: tcpID.DstPort,
		ServerIP:   tcpID.SrcIP,
		ServerPort: tcpID.SrcPort,
		IsOutgoing: true,
	}

	item := &api.OutputChannelItem{
		Protocol:       _protocol,
		Capture:        capture,
		Timestamp:      reqResPair.Request.CaptureTime.UnixNano() / int64(time.Millisecond),
		ConnectionInfo: connectionInfo,
		Pair: &api.RequestResponsePair{
			Request: api.GenericMessage{
				IsRequest:   true,
				CaptureTime: reqResPair.Request.CaptureTime,
				CaptureSize: int(reqResPair.Request.Size),
				Payload: KafkaPayload{
					Data: &KafkaWrapper{
						Method:  apiNames[apiKey],
						Url:     "",
						Details: reqResPair.Request,
					},
				},
			},
			Response: api.GenericMessage{
				IsRequest:   false,
				CaptureTime: reqResPair.Response.CaptureTime,
				CaptureSize: int(reqResPair.Response.Size),
				Payload: KafkaPayload{
					Data: &KafkaWrapper{
						Method:  apiNames[apiKey],
						Url:     "",
						Details: reqResPair.Response,
					},
				},
			},
		},
	}
	emitter.Emit(item)

	if i := int(apiKey); i < 0 || i >= numApis {
		err = fmt.Errorf("unsupported api key: %d", i)
		return err
	}

	d.discardAll()

	return nil
}
