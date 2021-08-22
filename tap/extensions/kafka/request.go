package main

import (
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/up9inc/mizu/tap/api"
)

type Request struct {
	Size          int32
	ApiKey        ApiKey
	ApiVersion    int16
	CorrelationID int32
	ClientID      string
	Payload       interface{}
}

func (req *Request) print() {
	log.Printf("> Request [%d]\n", req.Size)
	log.Printf("ApiKey: %v\n", req.ApiKey)
	log.Printf("ApiVersion: %v\n", req.ApiVersion)
	log.Printf("CorrelationID: %v\n", req.CorrelationID)
	log.Printf("ClientID: %v\n", req.ClientID)
	log.Printf("Payload: %+v\n", req.Payload)
}

func ReadRequest(r io.Reader, tcpID *api.TcpID) (apiKey ApiKey, apiVersion int16, err error) {
	d := &decoder{reader: r, remain: 4}
	size := d.readInt32()

	if err = d.err; err != nil {
		err = dontExpectEOF(err)
		return
	}

	d.remain = int(size)
	apiKey = ApiKey(d.readInt16())
	apiVersion = d.readInt16()
	correlationID := d.readInt32()
	clientID := d.readString()

	if i := int(apiKey); i < 0 || i >= len(apiTypes) {
		err = fmt.Errorf("unsupported api key: %d", i)
		return
	}

	if err = d.err; err != nil {
		err = dontExpectEOF(err)
		return
	}

	t := &apiTypes[apiKey]
	if t == nil {
		err = fmt.Errorf("unsupported api: %s", apiNames[apiKey])
		return
	}

	var payload interface{}

	switch apiKey {
	case Metadata:
		var mt interface{}
		var metadataRequest interface{}
		if apiVersion >= 11 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV11{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV11{}
		} else if apiVersion >= 10 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV10{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV10{}
		} else if apiVersion >= 8 {
			types := makeTypes(reflect.TypeOf(&MetadataRequestV8{}).Elem())
			mt = types[0]
			metadataRequest = &MetadataRequestV8{}
		} else if apiVersion >= 4 {
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
		break
	case ApiVersions:
		var mt interface{}
		var apiVersionsRequest interface{}
		if apiVersion >= 3 {
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
		break
	case Produce:
		var mt interface{}
		var produceRequest interface{}
		if apiVersion >= 3 {
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
		break
	case Fetch:
		var mt interface{}
		var fetchRequest interface{}
		if apiVersion >= 11 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV11{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV11{}
		} else if apiVersion >= 9 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV9{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV9{}
		} else if apiVersion >= 7 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV7{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV7{}
		} else if apiVersion >= 5 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV5{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV5{}
		} else if apiVersion >= 4 {
			types := makeTypes(reflect.TypeOf(&FetchRequestV4{}).Elem())
			mt = types[0]
			fetchRequest = &FetchRequestV4{}
		} else if apiVersion >= 3 {
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
		if apiVersion >= 4 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV4{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV4{}
		} else if apiVersion >= 2 {
			types := makeTypes(reflect.TypeOf(&ListOffsetsRequestV2{}).Elem())
			mt = types[0]
			listOffsetsRequest = &ListOffsetsRequestV2{}
		} else if apiVersion >= 1 {
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
		break
	case CreateTopics:
		var mt interface{}
		var createTopicsRequest interface{}
		if apiVersion >= 1 {
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
		break
	case DeleteTopics:
		var mt interface{}
		var deleteTopicsRequest interface{}
		if apiVersion >= 6 {
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
		log.Printf("[WARNING] (Request) Not implemented: %s\n", apiKey)
		break
	}

	request := &Request{
		Size:          size,
		ApiKey:        apiKey,
		ApiVersion:    apiVersion,
		CorrelationID: correlationID,
		ClientID:      clientID,
		Payload:       payload,
	}

	key := fmt.Sprintf(
		"%s:%s->%s:%s::%d",
		tcpID.SrcIP,
		tcpID.SrcPort,
		tcpID.DstIP,
		tcpID.DstPort,
		correlationID,
	)
	// fmt.Printf("key: %v\n", key)
	reqResMatcher.registerRequest(key, request)

	d.discardAll()

	return
}

func WriteRequest(w io.Writer, apiVersion int16, correlationID int32, clientID string, msg Message) error {
	apiKey := msg.ApiKey()

	if i := int(apiKey); i < 0 || i >= len(apiTypes) {
		return fmt.Errorf("unsupported api key: %d", i)
	}

	t := &apiTypes[apiKey]
	if t == nil {
		return fmt.Errorf("unsupported api: %s", apiNames[apiKey])
	}

	minVersion := t.minVersion()
	maxVersion := t.maxVersion()

	if apiVersion < minVersion || apiVersion > maxVersion {
		return fmt.Errorf("unsupported %s version: v%d not in range v%d-v%d", apiKey, apiVersion, minVersion, maxVersion)
	}

	r := &t.requests[apiVersion-minVersion]
	v := valueOf(msg)
	b := newPageBuffer()
	defer b.unref()

	e := &encoder{writer: b}
	e.writeInt32(0) // placeholder for the request size
	e.writeInt16(int16(apiKey))
	e.writeInt16(apiVersion)
	e.writeInt32(correlationID)

	if r.flexible {
		// Flexible messages use a nullable string for the client ID, then extra space for a
		// tag buffer, which begins with a size value. Since we're not writing any fields into the
		// latter, we can just write zero for now.
		//
		// See
		// https://cwiki.apache.org/confluence/display/KAFKA/KIP-482%3A+The+Kafka+Protocol+should+Support+Optional+Tagged+Fields
		// for details.
		e.writeNullString(clientID)
		e.writeUnsignedVarInt(0)
	} else {
		// Technically, recent versions of kafka interpret this field as a nullable
		// string, however kafka 0.10 expected a non-nullable string and fails with
		// a NullPointerException when it receives a null client id.
		e.writeString(clientID)
	}
	r.encode(e, v)
	err := e.err

	if err == nil {
		size := packUint32(uint32(b.Size()) - 4)
		b.WriteAt(size[:], 0)
		_, err = b.WriteTo(w)
	}

	return err
}
