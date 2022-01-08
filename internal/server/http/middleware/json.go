package middleware

import "github.com/gofiber/fiber/v2"

func JsonHeaders(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")

	return c.Next()
}