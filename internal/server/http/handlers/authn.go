package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"

	"github.com/Jameslikestea/grm/internal/authn"
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

		user, err := a.Username(token)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return nil
		}

		keys, err := stor.GetHashKey()
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return nil
		}

		as := storage.AuthenticationSession{
			User:      user,
			TID:       gocql.TimeUUID(),
			KID:       keys[0].KID,
			Type:      "USER",
			Signature: "",
		}

		hasher := sha512.New()
		hasher.Write([]byte(keys[0].K))
		as.Signature = hex.EncodeToString(hasher.Sum([]byte(as.UnhashedString())))

		ctx.Status(http.StatusOK)
		ctx.Write([]byte(as.String()))

		return nil
	}
}
