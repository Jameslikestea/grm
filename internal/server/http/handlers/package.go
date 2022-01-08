package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Package(ctx *fiber.Ctx) error {
	log.Info().Msg("serving package redirect")

	q := ctx.Query("q")
	return ctx.Redirect(fmt.Sprintf("/%s", q), http.StatusFound)
}
