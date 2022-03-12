package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/policy"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
)

func CreateNamespace(n namespace.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ns := ctx.Params("namespace")
		pub := ctx.Query("public", "1")
		uid := ctx.Locals(middleware.USER_ID).(string)

		log.Info().Str("namespace", ns).Str("public", pub).Str("uid", uid).Msg("Create ns")

		cns := models.CreateNamespaceRequest{
			Name:   ns,
			Public: pub != "0",
		}

		namespace, _ := n.GetNamespace(ns)
		if !p.Evaluate(
			policy.NamespaceCreate, policy.PolicyRequest{
				UserID:          uid,
				Namespace:       namespace,
				CreateNamespace: cns,
			},
		) {
			ctx.Status(http.StatusConflict)
			ctx.Write([]byte(http.StatusText(http.StatusConflict)))
			return nil
		}

		name := n.CreateNamespace(cns)
		n.CreateNamespacePermission(
			models.NamespacePermission{
				Namespace: name.Name,
				UserID:    uid,
				Admin:     true,
				Write:     true,
				Read:      true,
			},
		)
		ctx.JSON(name)

		return nil
	}
}
