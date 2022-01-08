package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Index(ctx *fiber.Ctx) error {
	log.Info().Msg("serving index")
	return ctx.Render(
		"index", nil,
	)
}
