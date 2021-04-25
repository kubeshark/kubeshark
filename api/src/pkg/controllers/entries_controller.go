package controllers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"mizuserver/src/pkg/database"
	"mizuserver/src/pkg/models"
)

func GetEntries(c *fiber.Ctx) error {
	var baseEntries []models.BaseEntryDetails
	database.EntriesCollection.Find(&baseEntries)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"entries": baseEntries,
	})
}

func GetEntry(c *fiber.Ctx) error {

	var entryData models.EntryData
	database.EntriesCollection.
		Select("entry").
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	var fullEntry har.Entry
	entryObject := json.Unmarshal([]byte(entryData.Entry), &fullEntry)


	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		 "msg":   entryObject,
	})
}
