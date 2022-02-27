package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
	"github.com/Jameslikestea/grm/internal/storage"
)

func HandleStartAuthenticator(a authn.Authenticator) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		u := a.NewSession()
		ctx.Redirect(u, http.StatusTemporaryRedirect)

		ctx.Write([]byte(u))
		return nil
	}
}

func HandleGithubAuthentication(a authn.Authenticator, stor storage.Storage) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		code := ctx.Query("code", "")
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
