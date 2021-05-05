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
func getSortAndOrder(operator string) (string, string) {
	var sortingOperator, ordering string
	if strings.ToLower(operator) == "gt" {
		sortingOperator = ">"
		ordering = "asc"
	} else if strings.ToLower(operator) == "lt" {
		sortingOperator = "<"
		ordering = "desc"
	} else {
		fmt.Println("Unsupported sort option")
		return "", ""
	}
	return sortingOperator, ordering

}
func GetEntries(c *fiber.Ctx) error {
	limit, e := strconv.Atoi(c.Query("limit", "200"));
	utils.CheckErr(e)
	if limit > HardLimit {
		fmt.Printf("Limit is greater than hard limit - using hard limit, requestedLimit: %v, hard: %v", limit ,HardLimit)
		limit = HardLimit
	}


	sortingOperator, ordering := getSortAndOrder(c.Query("operator", "lt"))
	timestamp, e := strconv.Atoi(c.Query("timestamp", "-1"))
	utils.CheckErr(e)

	var entries []models.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", ordering)).
		Where(fmt.Sprintf("timestamp %s %v",sortingOperator, timestamp)).
		Omit("entry"). // remove the "big" entry field
		Limit(limit).
		Find(&entries)

	if len(entries) > 0 && ordering == "desc"{
		utils.ReverseSlice(entries)
	}

	// Convert to base entries
	baseEntries := make([]models.BaseEntryDetails, 0, limit)
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

func GetGeneralStats(c *fiber.Ctx) error {
	sqlQuery := "SELECT count(*) as count, min(timestamp) as min, max(timestamp) as max from mizu_entries"
	var result struct {
		Count int
		Min int
		Max int
	}
	database.GetEntriesTable().Raw(sqlQuery).Scan(&result)
	return c.Status(fiber.StatusOK).JSON(&result)
}
