package controllers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"mizuserver/src/pkg/utils"
	"time"
)

type BaseEntryDetails struct {
	Id         string `json:"id,omitempty"`
	Url        string `json:"url,omitempty"`
	Service    string `json:"service,omitempty"`
	Path       string `json:"path,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Method     string `json:"method,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

var (
	entries [100]BaseEntryDetails

)

func GenerateData() {
	for i := 0; i < 100; i++ {
		entries[i] = BaseEntryDetails{
			Id:         primitive.NewObjectID().Hex(),
			Url:        utils.GetRandomString(10),
			Service:    utils.GetRandomString(10),
			Path:       utils.GetRandomString(10),
			StatusCode: rand.Int(),
			Method:     utils.GetRandomString(4),
			Timestamp:  time.Now().Unix(),
		}
	}
}


func GetEntries(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"entries": entries,
	})
}

func GetEntry(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "GetSingleBook",
	})
}
