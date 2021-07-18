package controllers

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/holder"
)

func GetCurrentResolvingInformation(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(holder.GetResolver().GetMap())
}

