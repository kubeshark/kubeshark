package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/up9inc/mizu/shared"
)

var TapStatus shared.TapStatus

func GetTappingStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(TapStatus)
}
