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
	id, _ := strconv.Atoi(c.Param("id"))
	var entry tapApi.MizuEntry
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, id)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bytes, &entry); err != nil {
		panic(err)
	}

	extension := extensionsMap[entry.Protocol.Name]
	protocol, representation, bodySize, _ := extension.Dissector.Represent(entry.Protocol, entry.Request, entry.Response)

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
		Protocol:       protocol,
		Representation: string(representation),
		BodySize:       bodySize,
		Data:           entry,
		Rules:          rules,
		IsRulesEnabled: isRulesEnabled,
	})
}
