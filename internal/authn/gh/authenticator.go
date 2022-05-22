package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ authn.Authenticator = &GithubAuthenticator{}

type GithubAuthenticator struct {
	conf    *oauth2.Config
	storage storage.Storage
}

type User struct {
	Username string `json:"login"`
	ID       int    `json:"id"`
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
		storage: stor,
	}
}

func (gh *GithubAuthenticator) NewSession(state string) string {
	return gh.conf.AuthCodeURL(state)
}

func (gh *GithubAuthenticator) UID(token string) (string, error) {
	client := gh.conf.Client(context.Background(), &oauth2.Token{AccessToken: token})
	response, err := client.Get("https://api.github.com/user")
	var m User
	err = json.NewDecoder(response.Body).Decode(&m)
	log.Info().Int("userid", m.ID).Msg("got UID")
	return fmt.Sprintf("GH:%X", m.ID), err
}

func (gh *GithubAuthenticator) LookupUID(username string) (string, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	response, err := client.Get(fmt.Sprintf("https://api.github.com/users/%s", username))
	var m User
	err = json.NewDecoder(response.Body).Decode(&m)
	return fmt.Sprintf("GH:%X", m.ID), err
}

func (gh *GithubAuthenticator) GetUsername(uid string) (string, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	id := strings.ReplaceAll(uid, "GH:", "")
	iid, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return "", err
	}

	response, err := client.Get(fmt.Sprintf("https://api.github.com/user/%d", iid))
	if err != nil {
		return "", err
	}
	var m User
	err = json.NewDecoder(response.Body).Decode(&m)
	return m.Username, err
}

func (gh *GithubAuthenticator) Register(token string) (string, error) {
	uid, err := gh.UID(token)
	if err != nil {
		return "", err
	}

	h := plumbing.ComputeHash(5, []byte(uid))

	user := models.User{UID: uid, Hash: h}

	err = gh.storage.StoreObject(
		"_internal._authn", storage.Object{
			Hash:    user.Hash,
			Type:    0,
			Content: models.Marshal(user),
		},
		0,
	)

	if err != nil {
		return "", err
	}

	return gh.CreateSession(user)
}

func (gh *GithubAuthenticator) CreateSession(userHash models.User) (string, error) {
	expiry := time.Now().UTC().Add(time.Hour)

	h := plumbing.ComputeHash(5, []byte(fmt.Sprintf("%s:%s", userHash.Hash.String(), expiry.Format(time.RFC3339))))

	session := models.UserSession{
		Hash:    h,
		User:    userHash,
		Expires: expiry,
	}

	err := gh.storage.StoreObject(
		"_internal._sessions", storage.Object{
			Hash:    session.Hash,
			Type:    0,
			Content: models.Marshal(session),
		},
		60*60,
	)

	if err != nil {
		return "", err
	}

	return session.Hash.String(), nil
}

func (gh *GithubAuthenticator) Token(code string) (string, error) {
	tok, err := gh.conf.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}

	return tok.AccessToken, nil
}

func (gh *GithubAuthenticator) GetSession(hash plumbing.Hash) (models.UserSession, error) {
	var u models.UserSession

	obj, err := gh.storage.GetObject("_internal._sessions", hash)
	if err != nil {
		return u, err
	}

	err = json.Unmarshal(obj.Content, &u)
	return u, err
}
