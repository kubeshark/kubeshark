package controllers

import (
	"github.com/gofiber/fiber/v2"
)

type VersionResponse struct {
	SemVer string `json:"semver"`
}

func GetVersion(c *fiber.Ctx) error {
	resp := VersionResponse{SemVer: "1.2.3"}
	return c.Status(fiber.StatusOK).JSON(resp)
}
