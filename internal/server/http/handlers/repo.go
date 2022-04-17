package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/policy"
	"github.com/Jameslikestea/grm/internal/repository"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
)

func Repository(c *fiber.Ctx) error {
	log.Info().Msg("serving repository")

	param := c.Params("*", "No Value Received")
	return c.SendString(fmt.Sprintf("%s/%s", config.GetDomain(), param))
}

func CreateRepository(n namespace.Manager, r repository.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ns := ctx.Params("namespace")
		repo := ctx.Params("repo")
		pub := ctx.Query("public", "1")
		uid := ctx.Locals(middleware.USER_ID).(string)

		log.Info().Str("namespace", ns).Str("public", pub).Str("uid", uid).Msg("Create ns")

		cns := models.CreateRepoRequest{
			Name:      repo,
			Namespace: ns,
			Public:    pub != "0",
		}

		namespace, _ := r.GetRepo(ns, repo)
		if !p.Evaluate(
			policy.RepoCreate, policy.PolicyRequest{
				UserID:     uid,
				Repo:       namespace,
				CreateRepo: cns,
			},
		) {
			ctx.Status(http.StatusConflict)
			ctx.Write([]byte(http.StatusText(http.StatusConflict)))
			return nil
		}

		name := r.CreateRepo(cns)
		nsp := models.RepoPermission{
			RepoName: name.Name,
			UserID:   uid,
			Admin:    true,
			Write:    true,
			Read:     true,
		}
		r.CreateRepoUserPermission(nsp)
		r.UpdateRepoPermissions(nsp)
		ctx.JSON(name)

		return nil
	}
}

func GetRepository(n namespace.Manager, r repository.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ns := ctx.Params("namespace")
		repo := ctx.Params("repo")
		pub := ctx.Query("public", "1")
		uid := ctx.Locals(middleware.USER_ID).(string)

		log.Info().Str("namespace", ns).Str("public", pub).Str("uid", uid).Msg("Create ns")

		perms := r.GetRepoPermissions(ns, repo)
		nsPerms := n.GetNamespacePermissions(ns)
		namespace, err := r.GetRepo(ns, repo)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			ctx.Write([]byte(http.StatusText(http.StatusNotFound)))
			return nil
		}

		allow := p.Evaluate(
			policy.RepoRead, policy.PolicyRequest{
				UserID:               uid,
				Repo:                 namespace,
				RepoPermissions:      perms,
				NamespacePermissions: nsPerms,
			},
		)

		log.Info().Bool("allow", allow).Str("namespace", ns).Str("repo", repo).Str(
			"user_id",
			uid,
		).Msg("User Requested Repo")

		if !allow {
			ctx.Status(http.StatusForbidden)
			ctx.Write([]byte(http.StatusText(http.StatusForbidden)))
			return nil
		}

		ctx.JSON(namespace)

		return nil
	}
}

func FERepository(n namespace.Manager, r repository.Manager, p policy.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ns := ctx.Params("namespace")
		repo := ctx.Params("repo")
		pub := ctx.Query("public", "1")
		uid := ctx.Locals(middleware.USER_ID).(string)
		auth := ctx.Locals(middleware.AUTHENTICATED).(bool)

		log.Info().Str("namespace", ns).Str("public", pub).Str("uid", uid).Msg("Create ns")

		perms := r.GetRepoPermissions(ns, repo)
		nsPerms := n.GetNamespacePermissions(ns)
		tags := r.GetTags(ns, repo)
		namespace, err := r.GetRepo(ns, repo)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			ctx.Write([]byte(http.StatusText(http.StatusNotFound)))
			return nil
		}

		allow := p.Evaluate(
			policy.RepoRead, policy.PolicyRequest{
				UserID:               uid,
				Repo:                 namespace,
				RepoPermissions:      perms,
				NamespacePermissions: nsPerms,
			},
		)

		log.Info().Bool("allow", allow).Str("namespace", ns).Str("repo", repo).Str(
			"user_id",
			uid,
		).Msg("User Requested Repo")

		if !allow {
			ctx.Status(http.StatusForbidden)
			ctx.Write([]byte(http.StatusText(http.StatusForbidden)))
			return nil
		}

		ctx.Render(
			"repository", fiber.Map{
				"Base":      config.GetDomain(),
				"Name":      namespace.Name,
				"Namespace": namespace.Namespace,
				"Tags":      tags,
				"Anon":      !auth,
			},
		)

		return nil
	}
}
