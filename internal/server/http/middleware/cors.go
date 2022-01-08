package middleware

import "github.com/gofiber/fiber/v2"

func Cors(c *fiber.Ctx) error {
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Headers", "*")
	c.Set("Access-Control-Allow-Methods", "GET, POST, HEAD, DELETE, OPTIONS")

	return c.Next()
}
