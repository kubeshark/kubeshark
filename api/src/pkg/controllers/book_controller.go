package controllers

import "github.com/gofiber/fiber/v2"

func GetBooks(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "GetBooks",
	})
}

func GetBook(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "GetSingleBook",
	})
}
