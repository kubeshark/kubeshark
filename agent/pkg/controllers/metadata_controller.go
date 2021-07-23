package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/version"
)

func GetVersion(c *fiber.Ctx) error {
	resp := shared.VersionResponse{SemVer: version.SemVer}
	return c.Status(fiber.StatusOK).JSON(resp)
}
