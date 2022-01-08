package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
)

func Repository(c *fiber.Ctx) error {
	log.Info().Msg("serving repository")

	param := c.Params("*", "No Value Received")
	return c.SendString(fmt.Sprintf("%s/%s", config.GetDomain(), param))
}
