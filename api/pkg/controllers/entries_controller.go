package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"strings"
	"time"
)

func GetEntries(c *fiber.Ctx) error {
	entriesFilter := &models.EntriesFilter{}

	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	order := database.OperatorToOrderMapping[entriesFilter.Operator]
	operatorSymbol := database.OperatorToSymbolMapping[entriesFilter.Operator]
	var entries []models.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", order)).
		Where(fmt.Sprintf("timestamp %s %v", operatorSymbol, entriesFilter.Timestamp)).
		Omit("entry"). // remove the "big" entry field
		Limit(entriesFilter.Limit).
		Find(&entries)

	if len(entries) > 0 && order == database.OrderDesc {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	baseEntries := make([]models.BaseEntryDetails, 0)
	for _, data := range entries {
		harEntry := models.BaseEntryDetails{}
		if err := models.GetEntry(&data, &harEntry); err != nil {
			continue
		}
		baseEntries = append(baseEntries, harEntry)
	}

	return c.Status(fiber.StatusOK).JSON(baseEntries)
}

func GetHARs(c *fiber.Ctx) error {
	entriesFilter := &models.HarFetchRequestBody{}
	order := database.OrderDesc
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
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

	var entries []models.MizuEntry
	database.GetEntriesTable().
		Where(fmt.Sprintf("timestamp BETWEEN %v AND %v", timestampFrom, timestampTo)).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	harsObject := map[string]*models.ExtendedHAR{}

	for _, entryData := range entries {
		var harEntry har.Entry
		_ = json.Unmarshal([]byte(entryData.Entry), &harEntry)
		if entryData.ResolvedDestination != "" {
			harEntry.Request.URL = utils.SetHostname(harEntry.Request.URL, entryData.ResolvedDestination)
		}

		var fileName string
		sourceOfEntry := entryData.ResolvedSource
		if sourceOfEntry != "" {
			// naively assumes the proper service source is http
			sourceOfEntry = fmt.Sprintf("http://%s", sourceOfEntry)
			//replace / from the file name cause they end up creating a corrupted folder
			fileName = fmt.Sprintf("%s.har", strings.ReplaceAll(sourceOfEntry, "/", "_"))
		} else {
			fileName = "unknown_source.har"
		}
		if harOfSource, ok := harsObject[fileName]; ok {
			harOfSource.Log.Entries = append(harOfSource.Log.Entries, &harEntry)
		} else {
			var entriesHar []*har.Entry
			entriesHar = append(entriesHar, &harEntry)
			harsObject[fileName] = &models.ExtendedHAR{
				Log: &models.ExtendedLog{
					Version: "1.2",
					Creator: &models.ExtendedCreator{
						Creator: &har.Creator{
							Name:    "mizu",
							Version: "0.0.2",
						},
					},
					Entries: entriesHar,
				},
			}
			// leave undefined when no source is present, otherwise modeler assumes source is empty string ""
			if sourceOfEntry != "" {
				harsObject[fileName].Log.Creator.Source = &sourceOfEntry
			}
		}
	}

	retObj := map[string][]byte{}
	for k, v := range harsObject {
		bytesData, _ := json.Marshal(v)
		retObj[k] = bytesData
	}
	buffer := utils.ZipData(retObj)
	return c.Status(fiber.StatusOK).SendStream(buffer)
}

func UploadEntries(c *fiber.Ctx) error {
	uploadRequestBody := &models.UploadEntriesRequestBody{}
	if err := c.QueryParser(uploadRequestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	if err := validation.Validate(uploadRequestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	if up9.GetAnalyzeInfo().IsAnalyzing {
		return c.Status(fiber.StatusBadRequest).SendString("Cannot analyze, mizu is already analyzing")
	}

	token, _ := up9.CreateAnonymousToken(uploadRequestBody.Dest)
	go up9.UploadEntriesImpl(token.Token, token.Model, uploadRequestBody.Dest)
	return c.Status(fiber.StatusOK).SendString("OK")
}

func GetFullEntries(c *fiber.Ctx) error {
	entriesFilter := &models.HarFetchRequestBody{}
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
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

	entriesArray := database.GetEntriesFromDb(timestampFrom, timestampTo)
	result := make([]models.FullEntryDetails, 0)
	for _, data := range entriesArray {
		harEntry := models.FullEntryDetails{}
		if err := models.GetEntry(&data, &harEntry); err != nil {
			continue
		}
		result = append(result, harEntry)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func GetEntry(c *fiber.Ctx) error {
	var entryData models.MizuEntry
	database.GetEntriesTable().
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	fullEntry := models.FullEntryDetails{}
	if err := models.GetEntry(&entryData, &fullEntry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Can't get entry details",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fullEntry)
}

func DeleteAllEntries(c *fiber.Ctx) error {
	database.GetEntriesTable().
		Where("1 = 1").
		Delete(&models.MizuEntry{})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": "Success",
	})

}

func GetGeneralStats(c *fiber.Ctx) error {
	sqlQuery := "SELECT count(*) as count, min(timestamp) as min, max(timestamp) as max from mizu_entries"
	var result struct {
		Count int
		Min   int
		Max   int
	}
	database.GetEntriesTable().Raw(sqlQuery).Scan(&result)
	return c.Status(fiber.StatusOK).JSON(&result)
}
