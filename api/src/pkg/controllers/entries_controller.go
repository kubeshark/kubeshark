package controllers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"mizuserver/src/pkg/database"
	"mizuserver/src/pkg/models"
	"mizuserver/src/pkg/utils"
)

func GetEntries(c *fiber.Ctx) error {
	var baseEntries []models.BaseEntryDetails
	database.EntriesTable.
		Find(&baseEntries).
		Limit(100)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"entries": baseEntries,
	})
}

func GetEntry(c *fiber.Ctx) error {

	var entryData models.EntryData
	database.EntriesTable.
		Select("entry").
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	// TODO: check why don't we get entry here
	var fullEntry har.Entry
	unmarshallErr := json.Unmarshal([]byte(entryData.Entry), &fullEntry)
	utils.CheckErr(unmarshallErr)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   fullEntry,
	})
}
