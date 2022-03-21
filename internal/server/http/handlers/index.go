package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/server/http/middleware"
)

func Index(ctx *fiber.Ctx) error {
	log.Info().Msg("serving index")
	auth := ctx.Locals(middleware.AUTHENTICATED).(bool)

	return ctx.Render(
		"index", fiber.Map{
			"Anon": !auth,
		},
	)
}
