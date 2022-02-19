package gh

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/Jameslikestea/grm/internal/authn"
)

var _ authn.Authenticator = &GithubAuthenticator{}

type GithubAuthenticator struct {
	conf *oauth2.Config
}

func New(
	ClientID string,
	ClientSecret string,
	RedirectURL string,
) *GithubAuthenticator {
	return &GithubAuthenticator{
		conf: &oauth2.Config{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  RedirectURL,
			Scopes: []string{
				"read:user",
			},
		},
	}
}

func (gh *GithubAuthenticator) Username(token string) (string, error) {
	return "", nil
}

func (gh *GithubAuthenticator) Register(token string) (string, error) {
	return "", nil
}

func (gh *GithubAuthenticator) Login(token string) (string, error) {
	return "", nil
}

func (gh *GithubAuthenticator) Token(code string) (string, error) {
	return "", nil
}
