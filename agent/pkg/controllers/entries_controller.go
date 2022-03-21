package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/validation"

	"github.com/gin-gonic/gin"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

func Error(c *gin.Context, err error) bool {
	if err != nil {
		logger.Log.Errorf("Error getting entry: %v", err)
		_ = c.Error(err)
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

	if entriesRequest.TimeoutMs == 0 {
		entriesRequest.TimeoutMs = 3000
	}

	data, meta, err := basenine.Fetch(shared.BasenineHost, shared.BaseninePort,
		entriesRequest.LeftOff, entriesRequest.Direction, entriesRequest.Query,
		entriesRequest.Limit, time.Duration(entriesRequest.TimeoutMs)*time.Millisecond)
	if err != nil {
		c.JSON(http.StatusInternalServerError, validationError)
	}

	response := &models.EntriesResponse{}
	var dataSlice []interface{}

	for _, row := range data {
		var entry *tapApi.Entry
		err = json.Unmarshal(row, &entry)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     true,
				"type":      "error",
				"autoClose": "5000",
				"msg":       string(row),
			})
			return // exit
		}

		extension := app.ExtensionsMap[entry.Protocol.Name]
		base := extension.Dissector.Summarize(entry)

		dataSlice = append(dataSlice, base)
	}

	var metadata *basenine.Metadata
	err = json.Unmarshal(meta, &metadata)
	if err != nil {
		logger.Log.Debugf("Error recieving metadata: %v", err.Error())
	}

	response.Data = dataSlice
	response.Meta = metadata

	c.JSON(http.StatusOK, response)
}

func GetEntry(c *gin.Context) {
	singleEntryRequest := &models.SingleEntryRequest{}

	if err := c.BindQuery(singleEntryRequest); err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	validationError := validation.Validate(singleEntryRequest)
	if validationError != nil {
		c.JSON(http.StatusBadRequest, validationError)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var entry *tapApi.Entry
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, id, singleEntryRequest.Query)
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

	extension := app.ExtensionsMap[entry.Protocol.Name]
	base := extension.Dissector.Summarize(entry)
	representation, bodySize, _ := extension.Dissector.Represent(entry.Request, entry.Response)

	var rules []map[string]interface{}
	var isRulesEnabled bool
	if entry.Protocol.Name == "http" {
		harEntry, _ := har.NewEntry(entry.Request, entry.Response, entry.StartTime, entry.ElapsedTime)
		_, rulesMatched, _isRulesEnabled := models.RunValidationRulesState(*harEntry, entry.Destination.Name)
		isRulesEnabled = _isRulesEnabled
		inrec, _ := json.Marshal(rulesMatched)
		if err := json.Unmarshal(inrec, &rules); err != nil {
			logger.Log.Error(err)
		}
	}

	c.JSON(http.StatusOK, tapApi.EntryWrapper{
		Protocol:       entry.Protocol,
		Representation: string(representation),
		BodySize:       bodySize,
		Data:           entry,
		Base:           base,
		Rules:          rules,
		IsRulesEnabled: isRulesEnabled,
	})
}
