package controllers

import (
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
		baseEntries = append(baseEntries, models.BaseEntryDetails{
			Id:              entry.EntryId,
			Url:             entry.Url,
			Service:         entry.Service,
			Path:            entry.Path,
			StatusCode:      entry.Status,
			Method:          entry.Method,
			Timestamp:       entry.Timestamp,
			RequestSenderIp: entry.RequestSenderIp,
		})
	}

	return c.Status(fiber.StatusOK).JSON(baseEntries)
}

func GetEntry(c *fiber.Ctx) error {
	var entryData models.EntryData
	database.GetEntriesTable().
		Select("entry").
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	var fullEntry har.Entry
	unmarshallErr := json.Unmarshal([]byte(entryData.Entry), &fullEntry)
	utils.CheckErr(unmarshallErr)

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
