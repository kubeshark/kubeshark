package controllers

import (
	"encoding/json"
	"fmt"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"net/http"
	"time"

	"github.com/google/martian/har"

	"github.com/gin-gonic/gin"
	"github.com/romana/rlog"

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

func UploadEntries(c *gin.Context) {
	rlog.Infof("Upload entries - started\n")

	uploadParams := &models.UploadEntriesRequestQuery{}
	if err := c.BindQuery(uploadParams); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := validation.Validate(uploadParams); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if up9.GetAnalyzeInfo().IsAnalyzing {
		c.String(http.StatusBadRequest, "Cannot analyze, mizu is already analyzing")
		return
	}

	rlog.Infof("Upload entries - creating token. dest %s\n", uploadParams.Dest)
	token, err := up9.CreateAnonymousToken(uploadParams.Dest)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Cannot analyze, mizu is already analyzing")
		return
	}
	rlog.Infof("Upload entries - uploading. token: %s model: %s\n", token.Token, token.Model)
	go up9.UploadEntriesImpl(token.Token, token.Model, uploadParams.Dest, uploadParams.SleepIntervalSec)
	c.String(http.StatusOK, "OK")
}

func GetFullEntries(c *gin.Context) {
	entriesFilter := &models.HarFetchRequestQuery{}
	if err := c.BindQuery(entriesFilter); err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	var timestampFrom, timestampTo int64

	if entriesFilter.From < 0 {
		timestampFrom = 0
	} else {
		timestampFrom = entriesFilter.From
	}
	if entriesFilter.To <= 0 {
		timestampTo = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		timestampTo = entriesFilter.To
	}

	entriesArray := database.GetEntriesFromDb(timestampFrom, timestampTo, nil)

	result := make([]har.Entry, 0)
	for _, data := range entriesArray {
		var pair tapApi.RequestResponsePair
		if err := json.Unmarshal([]byte(data.Entry), &pair); err != nil {
			continue
		}
		harEntry, err := utils.NewEntry(&pair)
		if err != nil {
			continue
		}
		result = append(result, *harEntry)
	}

	c.JSON(http.StatusOK, result)
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

func DeleteAllEntries(c *gin.Context) {
	database.GetEntriesTable().
		Where("1 = 1").
		Delete(&tapApi.MizuEntry{})

	c.JSON(http.StatusOK, map[string]string{
		"msg": "Success",
	})

}

func GetGeneralStats(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetGeneralStats())
}

func GetTappingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, providers.TapStatus)
}

func AnalyzeInformation(c *gin.Context) {
	c.JSON(http.StatusOK, up9.GetAnalyzeInfo())
}

func GetRecentTLSLinks(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetAllRecentTLSAddresses())
}
