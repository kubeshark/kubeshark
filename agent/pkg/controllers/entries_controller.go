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
