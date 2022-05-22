package authn

import (
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/models"
)

type Authenticator interface {
	NewSession(state string) string
	Token(string) (string, error)
	UID(string) (string, error)
	LookupUID(string) (string, error)
	GetUsername(string) (string, error)

	Register(string) (string, error)
	CreateSession(user models.User) (string, error)
	GetSession(hash plumbing.Hash) (models.UserSession, error)
}
