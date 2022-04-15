package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/policy"
	"github.com/Jameslikestea/grm/internal/repository"
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
		nsp := models.NamespacePermission{
			Namespace: name.Name,
			UserID:    uid,
			Admin:     true,
			Write:     true,
			Read:      true,
		}
		n.CreateNamespaceUserPermission(nsp)
		n.UpdateNamespacePermissions(nsp)
		ctx.JSON(name)

		return nil
	}
}

func GetNamespace(n namespace.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ns := ctx.Params("namespace")
		uid := ctx.Locals(middleware.USER_ID).(string)

		perms := n.GetNamespacePermissions(ns)
		nspc, err := n.GetNamespace(ns)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			ctx.Write([]byte(http.StatusText(http.StatusNotFound)))
		}

		allow := p.Evaluate(
			policy.NamespaceRead, policy.PolicyRequest{
				UserID:               uid,
				NamespacePermissions: perms,
				Namespace:            nspc,
			},
		)

		log.Debug().Bool("allow", allow).Interface(
			"input",
			policy.PolicyRequest{UserID: uid, NamespacePermissions: perms, Namespace: nspc},
		).Msg("Made a decision")

		if !allow {
			ctx.Status(http.StatusForbidden)
			ctx.Write([]byte(http.StatusText(http.StatusForbidden)))
			return nil
		}

		ctx.JSON(nspc)

		return nil
	}
}

func FENamespace(n namespace.Manager, r repository.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		log.Info().Msg("rendering namespace")

		ns := ctx.Params("namespace")
		uid := ctx.Locals(middleware.USER_ID).(string)
		auth := ctx.Locals(middleware.AUTHENTICATED).(bool)

		perms := n.GetNamespacePermissions(ns)
		nspc, err := n.GetNamespace(ns)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			ctx.Render(
				"error", fiber.Map{
					"Code":    http.StatusNotFound,
					"Message": "Ooops... That namespace doesn't appear to exist...",
					"Anon":    !auth,
				},
			)
			return nil
		}

		repos, err := r.GetReposByNamespace(ns)
		if err != nil {
			log.Warn().Err(err).Str("namespace", ns).Msg("Cannot get repos for namespace")
		}

		var repoList []models.Repo
		for _, repo := range repos {
			rperm := r.GetRepoPermissions(ns, repo.Name)
			a := p.Evaluate(
				policy.RepoRead, policy.PolicyRequest{
					UserID:               uid,
					NamespacePermissions: perms,
					RepoPermissions:      rperm,
					Repo:                 repo,
				},
			)
			if a {
				repoList = append(repoList, repo)
			}
		}

		allow := p.Evaluate(
			policy.NamespaceRead, policy.PolicyRequest{
				UserID:               uid,
				NamespacePermissions: perms,
				Namespace:            nspc,
			},
		)

		if !allow {
			ctx.Status(http.StatusForbidden)
			ctx.Render(
				"error", fiber.Map{
					"Code":    http.StatusForbidden,
					"Message": "Ooops... You don't appear to have permission to see that...",
					"Anon":    !auth,
				},
			)
			return nil
		}

		ctx.Render(
			"namespace", fiber.Map{
				"Name":   nspc.Name,
				"Public": nspc.Public,
				"Repos":  repoList,
				"Anon":   !auth,
			},
		)
		return nil
	}
}
