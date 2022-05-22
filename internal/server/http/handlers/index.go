package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/policy"
	"github.com/Jameslikestea/grm/internal/repository"
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

func FEMe(n namespace.Manager, r repository.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		pub := ctx.Query("public", "1")
		uid := ctx.Locals(middleware.USER_ID).(string)
		auth := ctx.Locals(middleware.AUTHENTICATED).(bool)

		log.Info().Bool("authenticated", auth).Str("public", pub).Str("uid", uid).Msg("Create ns")

		namespaces, err := n.ListNamespaces()
		if err != nil {
			return err
		}

		nss := []models.Namespace{}

		for _, ns := range namespaces {
			perms := n.GetNamespacePermissions(ns.Name)

			if p.Evaluate(
				policy.NamespaceAdmin, policy.PolicyRequest{
					UserID:               uid,
					NamespacePermissions: perms,
					Namespace:            ns,
				},
			) {
				nss = append(nss, ns)
			}
		}

		ctx.Render(
			"me", fiber.Map{
				"Namespaces": nss,
				"Anon":       !auth,
			},
		)

		return nil
	}
}
