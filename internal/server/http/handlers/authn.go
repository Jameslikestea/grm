package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/pubkey"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
	"github.com/Jameslikestea/grm/internal/storage"
)

func HandleStartAuthenticator(a authn.Authenticator) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		u := a.NewSession(ctx.Get("Referer"))
		ctx.Redirect(u, http.StatusTemporaryRedirect)

		ctx.Write([]byte(u))
		return nil
	}
}

func HandleGithubAuthentication(a authn.Authenticator, stor storage.Storage) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		code := ctx.Query("code", "")
		state := ctx.Query("state", "")

		if code == "" {
			ctx.Status(http.StatusUnauthorized)
			ctx.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return nil
		}

		token, err := a.Token(code)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return nil
		}

		sess, err := a.Register(token)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return nil
		}

		ctx.Status(http.StatusOK)
		ctx.Cookie(
			&fiber.Cookie{
				Name:    "grm.authentication",
				Value:   sess,
				Expires: time.Now().UTC().Add(time.Hour),
			},
		)

		ctx.Redirect(state, http.StatusTemporaryRedirect)
		return nil
	}
}

func HandleMe(ctx *fiber.Ctx) error {

	ctx.Write(
		[]byte(fmt.Sprintf(
			"%v-%s-%s",
			ctx.Locals(middleware.AUTHENTICATED),
			ctx.Locals(middleware.USER_ID),
			ctx.Locals(middleware.SESSION_ID),
		)),
	)

	return nil
}

func HandleAddSSHKey(ps pubkey.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		uid := ctx.Locals(middleware.USER_ID).(string)
		auth := ctx.Locals(middleware.AUTHENTICATED).(bool)

		if !auth {
			ctx.Status(http.StatusForbidden)
			ctx.Write([]byte(http.StatusText(http.StatusForbidden)))
		}

		err := ps.StoreKey(
			models.UserPubKey{
				Key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDIbGv8RbhxbZ+UXkqkOmNOGCHvKqUcecuLZAeG6j3YDRJ8jnWaJMmfgsiUbot9Au2bqCItXsK0A027IjbY6QeMmZcSEUQfEbcc72/aWxYuyp5rT78lcF6ZkjTnAu6GOPfQG92uxkSHqvkpOpsoOw8dzOzVWeBz5aggS1B8yRQId5OwtAKz8BvmVuFocBlLZLaXAniGzwGqOeYaqIkDFyZxy9gG5J80fu61tBEDcb8TNdXg571oDaP48g+4r+X8SwnSQO3b7Bd8EYaZVdEE2g5qMOutl6ibLqYbUHpsNTjq88JiYQlC/yKWScCvOniaA4rKDNBN9asgN2gnlGcHjYWAclOc8zxoRXByOjQBBcJmHIr52MGRFZfGMYg0MuQlPXIEzTyHd23p3qgKuWD/kXpb6m20De02e75j/sBntAeGnjVYE6gctbHrRxV0lXOf2PF0XLZMVJswnJi1oxdNjqhiC/xJuUnrS60HIRmsN4KTObji+RLtIZkt9jF1kIpK27M= james@JC-LAPTOP",
				UID: uid,
			},
		)

		if err != nil {
			ctx.Status(http.StatusBadRequest)
			ctx.Write([]byte(fmt.Sprintf("error: %v", err)))
		}

		return nil
	}
}
