package controllers

import (
	"encoding/json"
	"fmt"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
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

func GetEntries(c *gin.Context) {
	entriesFilter := &models.EntriesFilter{}

	if err := c.BindQuery(entriesFilter); err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	order := database.OperatorToOrderMapping[entriesFilter.Operator]
	operatorSymbol := database.OperatorToSymbolMapping[entriesFilter.Operator]
	var entries []tapApi.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", order)).
		Where(fmt.Sprintf("timestamp %s %v", operatorSymbol, entriesFilter.Timestamp)).
		Limit(entriesFilter.Limit).
		Find(&entries)

	if len(entries) > 0 && order == database.OrderDesc {
		// the entries always order from oldest to newest - we should reverse
		utils.ReverseSlice(entries)
	}

	baseEntries := make([]tapApi.BaseEntryDetails, 0)
	for _, entry := range entries {
		baseEntryDetails := tapApi.BaseEntryDetails{}
		if err := models.GetEntry(&entry, &baseEntryDetails); err != nil {
			continue
		}

		var pair tapApi.RequestResponsePair
		json.Unmarshal([]byte(entry.Entry), &pair)
		harEntry, err := utils.NewEntry(&pair)
		if err == nil {
			rules, _, _ := models.RunValidationRulesState(*harEntry, entry.Service)
			baseEntryDetails.Rules = rules
		}

		baseEntries = append(baseEntries, baseEntryDetails)
	}

	c.JSON(http.StatusOK, baseEntries)
}

func GetEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("entryId"))
	fmt.Printf("GetEntry id: %v\n", id)
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
			Abbreviation:    entry["proto"].(map[string]interface{})["abbreviation"].(string),
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
		Path:            entry["path"].(string),
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
		var pair tapApi.RequestResponsePair
		json.Unmarshal([]byte(entryData.Entry), &pair)
		harEntry, _ := utils.NewEntry(&pair)
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
