package controllers

import (
	"encoding/json"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var extensionsMap map[string]*tapApi.Extension // global

func InitExtensionsMap(ref map[string]*tapApi.Extension) {
	extensionsMap = ref
}

func GetEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("entryId"))
	var entry map[string]interface{}
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, id)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bytes, &entry); err != nil {
		panic(err)
	}
	var response map[string]interface{}
	if entry["response"] != nil {
		response = entry["response"].(map[string]interface{})
	}
	entryData := tapApi.MizuEntry{
		Protocol: tapApi.Protocol{
			Name:            entry["proto"].(map[string]interface{})["name"].(string),
			LongName:        entry["proto"].(map[string]interface{})["longName"].(string),
			Abbreviation:    entry["proto"].(map[string]interface{})["abbr"].(string),
			Version:         entry["proto"].(map[string]interface{})["version"].(string),
			BackgroundColor: entry["proto"].(map[string]interface{})["backgroundColor"].(string),
			ForegroundColor: entry["proto"].(map[string]interface{})["foregroundColor"].(string),
			FontSize:        int8(entry["proto"].(map[string]interface{})["fontSize"].(float64)),
			ReferenceLink:   entry["proto"].(map[string]interface{})["referenceLink"].(string),
			Priority:        uint8(entry["proto"].(map[string]interface{})["priority"].(float64)),
		},
		Request:         entry["request"].(map[string]interface{}),
		Response:        response,
		EntryId:         entry["entryId"].(string),
		Entry:           entry["entry"].(string),
		Url:             entry["url"].(string),
		Method:          entry["method"].(string),
		Status:          int(entry["status"].(float64)),
		RequestSenderIp: entry["requestSenderIp"].(string),
		Service:         entry["service"].(string),
		Timestamp:       int64(entry["timestamp"].(float64)),
		ElapsedTime:     int64(entry["elapsedTime"].(float64)),
		Summary:         entry["summary"].(string),
		// ResolvedSource:      entry["resolvedSource"].(string),
		// ResolvedDestination: entry["resolvedDestination"].(string),
		SourceIp:        entry["sourceIp"].(string),
		DestinationIp:   entry["destinationIp"].(string),
		SourcePort:      entry["sourcePort"].(string),
		DestinationPort: entry["destinationPort"].(string),
		// IsOutgoing:      entry["isOutgoing"].(bool),
	}

	extension := extensionsMap[entryData.Protocol.Name]
	protocol, representation, bodySize, _ := extension.Dissector.Represent(&entryData)

	var rules []map[string]interface{}
	var isRulesEnabled bool
	if entryData.Protocol.Name == "http" {
		harEntry, _ := utils.NewEntry(entryData.Request, entryData.Response, entryData.StartTime, entryData.ElapsedTime)
		_, rulesMatched, _isRulesEnabled := models.RunValidationRulesState(*harEntry, entryData.Service)
		isRulesEnabled = _isRulesEnabled
		inrec, _ := json.Marshal(rulesMatched)
		json.Unmarshal(inrec, &rules)
	}

	c.JSON(http.StatusOK, tapApi.MizuEntryWrapper{
		Protocol:       protocol,
		Representation: string(representation),
		BodySize:       bodySize,
		Data:           entryData,
		Rules:          rules,
		IsRulesEnabled: isRulesEnabled,
	})
}
