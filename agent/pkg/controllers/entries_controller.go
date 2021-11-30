package controllers

import (
	"encoding/json"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var extensionsMap map[string]*tapApi.Extension // global

func InitExtensionsMap(ref map[string]*tapApi.Extension) {
	extensionsMap = ref
}

func Error(c *gin.Context, err error) bool {
	if err != nil {
		logger.Log.Errorf("Error getting entry: %v", err)
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":     true,
			"type":      "error",
			"autoClose": "5000",
			"msg":       err.Error(),
		})
		return true // signal that there was an error and the caller should return
	}
	return false // no error, can continue
}

func GetEntries(c *gin.Context) {
	entriesRequest := &models.EntriesRequest{}

	if err := c.BindQuery(entriesRequest); err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	validationError := validation.Validate(entriesRequest)
	if validationError != nil {
		c.JSON(http.StatusBadRequest, validationError)
	}

	data, meta, err := basenine.Fetch(shared.BasenineHost, shared.BaseninePort,
		entriesRequest.LeftOff, entriesRequest.Direction, entriesRequest.Query,
		entriesRequest.Limit, time.Duration(entriesRequest.TimeoutMs)*time.Millisecond)
	if err != nil {
		c.JSON(http.StatusInternalServerError, validationError)
	}

	result := make(map[string]interface{})
	var dataSlice []interface{}

	for _, row := range data {
		var dataMap map[string]interface{}
		err = json.Unmarshal(row, &dataMap)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     true,
				"type":      "error",
				"autoClose": "5000",
				"msg":       string(row),
			})
			return // exit
		}

		base := dataMap["base"].(map[string]interface{})
		base["id"] = uint(dataMap["id"].(float64))

		dataSlice = append(dataSlice, base)
	}

	var metadata *basenine.Metadata
	err = json.Unmarshal(meta, &metadata)
	if err != nil {
		logger.Log.Debugf("Error recieving metadata: %v", err.Error())
	}

	result["data"] = dataSlice
	result["meta"] = metadata

	c.JSON(http.StatusOK, result)
}

func GetEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry tapApi.MizuEntry
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, id)
	if Error(c, err) {
		return // exit
	}
	err = json.Unmarshal(bytes, &entry)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     true,
			"type":      "error",
			"autoClose": "5000",
			"msg":       string(bytes),
		})
		return // exit
	}

	extension := extensionsMap[entry.Protocol.Name]
	representation, bodySize, _ := extension.Dissector.Represent(entry.Request, entry.Response)

	var rules []map[string]interface{}
	var isRulesEnabled bool
	if entry.Protocol.Name == "http" {
		harEntry, _ := utils.NewEntry(entry.Request, entry.Response, entry.StartTime, entry.ElapsedTime)
		_, rulesMatched, _isRulesEnabled := models.RunValidationRulesState(*harEntry, entry.Service)
		isRulesEnabled = _isRulesEnabled
		inrec, _ := json.Marshal(rulesMatched)
		json.Unmarshal(inrec, &rules)
	}

	c.JSON(http.StatusOK, tapApi.MizuEntryWrapper{
		Protocol:       entry.Protocol,
		Representation: string(representation),
		BodySize:       bodySize,
		Data:           entry,
		Rules:          rules,
		IsRulesEnabled: isRulesEnabled,
	})
}
