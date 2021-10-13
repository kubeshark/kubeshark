package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	tapApi "github.com/up9inc/mizu/tap/api"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"net/http"
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
		Omit("entry"). // remove the "big" entry field
		Limit(entriesFilter.Limit).
		Find(&entries)

	if len(entries) > 0 && order == database.OrderDesc {
		// the entries always order from oldest to newest - we should reverse
		utils.ReverseSlice(entries)
	}

	baseEntries := make([]tapApi.BaseEntryDetails, 0)
	for _, data := range entries {
		harEntry := tapApi.BaseEntryDetails{}
		if err := models.GetEntry(&data, &harEntry); err != nil {
			continue
		}
		baseEntries = append(baseEntries, harEntry)
	}

	c.JSON(http.StatusOK, baseEntries)
}

func GetEntry(c *gin.Context) {
	var entryData tapApi.MizuEntry
	database.GetEntriesTable().
		Where(map[string]string{"entryId": c.Param("entryId")}).
		First(&entryData)

	extension := extensionsMap[entryData.ProtocolName]
	protocol, representation, bodySize, _ := extension.Dissector.Represent(&entryData)

	var rules []map[string]interface{}
	var isRulesEnabled bool
	if entryData.ProtocolName == "http" {
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
