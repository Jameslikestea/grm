package gh

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ authn.Authenticator = &GithubAuthenticator{}

type GithubAuthenticator struct {
	conf *oauth2.Config
}

func New(
	ClientID string,
	ClientSecret string,
	RedirectURL string,
	stor storage.Storage,
) *GithubAuthenticator {
	err := stor.GenerateHashKey()
	if err != nil {
		log.Error().Err(err).Msg("cannot generate hash key")
	}
	return &GithubAuthenticator{
		conf: &oauth2.Config{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  RedirectURL + "/authn/github",
			Scopes: []string{
				"read:user",
			},
		},
	}
}

func (gh *GithubAuthenticator) NewSession() string {
	return gh.conf.AuthCodeURL("no-state")
}

func (gh *GithubAuthenticator) Username(token string) (string, error) {

	type User struct {
		Username string `json:"login"`
		ID       int    `json:"id"`
	}

	client := gh.conf.Client(context.Background(), &oauth2.Token{AccessToken: token})
	response, err := client.Get("https://api.github.com/user")
	var m User
	err = json.NewDecoder(response.Body).Decode(&m)
	log.Info().Int("userid", m.ID).Msg("got UID")
	return fmt.Sprintf("GH:%X", m.ID), err
}

func (gh *GithubAuthenticator) Register(token string) (string, error) {
	return "", nil
}

func (gh *GithubAuthenticator) Login(token string) (string, error) {
	return "", nil
}

func (gh *GithubAuthenticator) Token(code string) (string, error) {
	tok, err := gh.conf.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}

	return tok.AccessToken, nil
}
