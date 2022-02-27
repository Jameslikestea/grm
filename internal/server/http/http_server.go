package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/authn/gh"
	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/server/http/handlers"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
	"github.com/Jameslikestea/grm/internal/storage"
	"github.com/Jameslikestea/grm/internal/storage/cql"
	"github.com/Jameslikestea/grm/internal/storage/memory"
)

type Server struct {
	s     *fiber.App
	stor  storage.Storage
	authn authn.Authenticator
}

func NewServer() *Server {
	engine := html.New("./templates", ".html")

	var stor storage.Storage
	var authn authn.Authenticator

	switch strings.ToUpper(config.GetStorageType()) {
	case "MEMORY":
		stor = memory.NewMemoryStorage()
	// case "SQLITE":
	// 	stor = sqlite.NewSQLLiteStorage()
	// case "MYSQL":
	// 	stor = mysql.NewSQLLiteStorage()
	// case "POSTGRES":
	// 	stor = postgres.NewSQLLiteStorage()
	// case "S3":
	// 	stor = s3.NewS3Storage()
	case "CQL":
		stor = cql.NewCQLStorage()
	default:
		log.Warn().Msg("No Acceptable Storage Engine Chosen, Defaulting to In Memory")
		stor = memory.NewMemoryStorage()
	}

	switch config.GetAuthenticationProvider() {
	case "GITHUB":
		authn = gh.New(
			config.GetAuthenticationGithubClientID(),
			config.GetAuthenticationGithubClientSecret(),
			config.GetAuthenticationGithubRedirectURL(),
			stor,
		)
	default:
		authn = nil
	}

	s := &Server{
		s: fiber.New(
			fiber.Config{
				Views:                 engine,
				AppName:               "grmpkg",
				DisableStartupMessage: true,
			},
		),
		stor:  stor,
		authn: authn,
	}

	s.constructMiddleware()
	s.constructRoutes()

	return s
}

// constructRoutes adds in all of the specific and generic route handlers
func (s *Server) constructRoutes() {
	s.s.Get("/", handlers.Index)
	s.s.Get("/package", handlers.Package)
	s.s.Get("/*.git", handlers.Git)
	s.s.Get("/*.git/info/refs", handlers.AdvertiseReference(s.stor))
	s.s.Post("/*.git/git-upload-pack", handlers.UploadPack(s.stor))

	s.s.Get("/authn/start", handlers.HandleStartAuthenticator(s.authn))
	s.s.Get("/authn/github", handlers.HandleGithubAuthentication(s.authn, s.stor))

	s.s.Get("/*", handlers.Repository)

}

// constructMiddleware adds all of the middleware required into the server, this is all of the defaults
func (s *Server) constructMiddleware() {
	s.s.Use(
		middleware.Cors,
		middleware.JsonHeaders,
	)
}

// Test - Setting ourselves up to enable testing of the API in the future
func (s *Server) Test(req *http.Request) (*http.Response, error) {
	return s.s.Test(req)
}

func (s *Server) Run() {
	log.Info().Msg("Starting HTTP")
	s.s.Listen(fmt.Sprintf("%s:%s", config.GetHTTPInterface(), config.GetHTTPPort()))
}
