package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"strconv"
	"strings"
)

const (
	HardLimit = 200
)

func GetEntries(c *fiber.Ctx) error {
	limit, e := strconv.Atoi(c.Query("limit", "200"))
	utils.CheckErr(e)
	if limit > HardLimit {
		limit = HardLimit
	}

	sortOption := c.Query("operator", "lt")
	var sortingOperator string

	if strings.ToLower(sortOption) == "gt" {
		sortingOperator = ">"
	} else if strings.ToLower(sortOption) == "lt" {
		sortingOperator = "<"
	} else {
		fmt.Println("Unsupported")
		return nil
	}

	timestamp := c.Query("timestamp")
	var entries []models.MizuEntry
	database.GetEntriesTable().
		Where(fmt.Sprintf("timestamp %s %s",sortingOperator, timestamp)).
		Omit("entry"). // remove the "big" entry field
		Limit(limit).
		Find(&entries)

	//	// Convert to base entries
	baseEntries := make([]models.BaseEntryDetails, 0)
	for _, entry := range entries {
		baseEntries = append(baseEntries, models.BaseEntryDetails{
			Id:         entry.EntryId,
			Url:        entry.Url,
			Service:    entry.Service,
			Path:       entry.Path,
			StatusCode: entry.Status,
			Method:     entry.Method,
			Timestamp:  entry.Timestamp,
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
