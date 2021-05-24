package controllers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
)

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
	LT        = "lt"
	GT        = "gt"
)

var (
	operatorToSymbolMapping = map[string]string{
		LT: "<",
		GT: ">",
	}
	operatorToOrderMapping = map[string]string{
		LT: OrderDesc,
		GT: OrderAsc,
	}
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

	order := operatorToOrderMapping[entriesFilter.Operator]
	operatorSymbol := operatorToSymbolMapping[entriesFilter.Operator]
	var entries []models.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", order)).
		Where(fmt.Sprintf("timestamp %s %v", operatorSymbol, entriesFilter.Timestamp)).
		Omit("entry"). // remove the "big" entry field
		Limit(entriesFilter.Limit).
		Find(&entries)

	if len(entries) > 0 && order == OrderDesc {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	// Convert to base entries
	baseEntries := make([]models.BaseEntryDetails, 0, entriesFilter.Limit)
	for _, entry := range entries {
		baseEntries = append(baseEntries, utils.GetResolvedBaseEntry(entry))
	}

	return c.Status(fiber.StatusOK).JSON(baseEntries)
}

func GetHAR(c *fiber.Ctx) error {
	entriesFilter := &models.HarFetchRequestBody{}
	order := OrderDesc
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	var entries []models.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", order)).
		// Where(fmt.Sprintf("timestamp %s %v", operatorSymbol, entriesFilter.Timestamp)).
		Limit(1000).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	retObj := map[string]*har.HAR{}

	for _, entryData := range entries {
		harEntryObject := []byte(entryData.Entry)

		var harEntry har.Entry
		_ = json.Unmarshal(harEntryObject, &harEntry)

		sourceOfEntry := *entryData.ResolvedSource
		if harOfSource, ok := retObj[sourceOfEntry]; ok {
			harOfSource.Log.Entries = append(harOfSource.Log.Entries, &harEntry)
		} else {
			var entriesHar []*har.Entry
			entriesHar = append(entriesHar, &harEntry)
			retObj[sourceOfEntry] = &har.HAR{
				Log: &har.Log{
					Version: "1.2",
					Creator: &har.Creator{
						Name:    "mizu",
						Version: "0.0.1",
					},
					Entries: entriesHar,
				},
			}
		}
	}

	buffer := zipData(retObj)
	return c.Status(fiber.StatusOK).SendStream(buffer)
}

func zipData(files map[string]*har.HAR) *bytes.Buffer {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)
	// Create a new zip archive.
	zipWriter := zip.NewWriter(buf)
	defer func() { _ = zipWriter.Close() }()

	for fileName, fileData := range files {
		zipFile, _ := zipWriter.Create(fileName + ".json")
		fileDataBytes, _ := json.Marshal(fileData)
		_, _ = zipFile.Write(fileDataBytes)
	}

	return buf
}

func GetEntry(c *fiber.Ctx) error {
	var entryData models.EntryData
	database.GetEntriesTable().
		Select("entry", "resolvedDestination").
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	var fullEntry har.Entry
	unmarshallErr := json.Unmarshal([]byte(entryData.Entry), &fullEntry)
	utils.CheckErr(unmarshallErr)

	if entryData.ResolvedDestination != nil {
		fullEntry.Request.URL = utils.SetHostname(fullEntry.Request.URL, *entryData.ResolvedDestination)
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
