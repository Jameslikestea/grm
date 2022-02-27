package middleware

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gofiber/fiber/v2"

	"github.com/Jameslikestea/grm/internal/authn"
)

const (
	AUTHENTICATED = "AUTHENTICATED"
	USER_ID       = "USER_ID"
	SESSION_ID    = "SESSION_ID"
)

func Authn(a authn.Authenticator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(AUTHENTICATED, false)
		c.Locals(USER_ID, "")
		c.Locals(SESSION_ID, "")

		h := c.Cookies("grm.authentication")

		if h == "" {
			return c.Next()
		}

		hash := plumbing.NewHash(h)
		if hash == plumbing.ZeroHash {
			return c.Next()
		}

		sess, err := a.GetSession(hash)
		if err != nil {
			return c.Next()
		}

		c.Locals(AUTHENTICATED, true)
		c.Locals(USER_ID, sess.User.UID)
		c.Locals(SESSION_ID, sess.Hash.String())

		return c.Next()
	}
}
