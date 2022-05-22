package http

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/authn"
	"github.com/Jameslikestea/grm/internal/authn/gh"
	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/namespace"
	servicens "github.com/Jameslikestea/grm/internal/namespace/service"
	"github.com/Jameslikestea/grm/internal/policy"
	"github.com/Jameslikestea/grm/internal/pubkey"
	serviceps "github.com/Jameslikestea/grm/internal/pubkey/service"
	"github.com/Jameslikestea/grm/internal/repository"
	servicers "github.com/Jameslikestea/grm/internal/repository/service"
	"github.com/Jameslikestea/grm/internal/server/http/handlers"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
	"github.com/Jameslikestea/grm/internal/storage"
	"github.com/Jameslikestea/grm/internal/storage/cql"
	"github.com/Jameslikestea/grm/internal/storage/memory"
	"github.com/Jameslikestea/grm/internal/storage/s3aws"
)

//go:embed templates
var _templates embed.FS

type Server struct {
	s     *fiber.App
	stor  storage.Storage
	authn authn.Authenticator
	pol   policy.Manager
	ns    namespace.Manager
	rs    repository.Manager
	ps    pubkey.Manager
}

func NewServer() *Server {
	fsys, _ := fs.Sub(_templates, "templates")
	engine := html.NewFileSystem(http.FS(fsys), ".html")
	engine.Load()

	var stor storage.Storage
	var authn authn.Authenticator
	pol := policy.New()

	switch strings.ToUpper(config.GetStorageType()) {
	case "MEMORY":
		stor = memory.NewMemoryStorage()
	// case "SQLITE":
	// 	stor = sqlite.NewSQLLiteStorage()
	// case "MYSQL":
	// 	stor = mysql.NewSQLLiteStorage()
	// case "POSTGRES":
	// 	stor = postgres.NewSQLLiteStorage()
	case "S3":
		stor = s3.NewS3Storage()
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

	ns := servicens.New(stor)
	rs := servicers.New(stor)
	ps := serviceps.New(stor)

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
		pol:   pol,
		ns:    ns,
		rs:    rs,
		ps:    ps,
	}

	s.constructMiddleware()
	s.constructRoutes()

	return s
}

// constructRoutes adds in all of the specific and generic route handlers
func (s *Server) constructRoutes() {
	s.s.Get("/", handlers.Index)

	s.s.Get(
		"/humans.txt", func(ctx *fiber.Ctx) error {
			ctx.Render("humans", fiber.Map{})
			return nil
		},
	)

	s.s.Get("/authn/start", handlers.HandleStartAuthenticator(s.authn))
	s.s.Get("/authn/github", handlers.HandleGithubAuthentication(s.authn, s.stor))
	s.s.Get("/authn/me", handlers.HandleMe)
	s.s.Post("/authn/ssh", handlers.HandleAddSSHKey(s.ps))

	s.s.Get("/*.git", handlers.Git)
	s.s.Get("/*.git/info/refs", handlers.AdvertiseReference(s.stor))
	s.s.Post("/*.git/git-upload-pack", handlers.UploadPack(s.stor))

	s.s.Get("/me", handlers.FEMe(s.ns, s.rs, s.pol))

	s.s.Get("/:namespace", handlers.FENamespace(s.ns, s.rs, s.pol, s.authn))
	s.s.Get("/:namespace/:repo", handlers.FERepository(s.ns, s.rs, s.pol, s.authn))
	s.s.Get("/:namespace/:repo/*", handlers.FERepository(s.ns, s.rs, s.pol, s.authn))
	s.s.Get("/package", handlers.Package)

	s.s.Post("/api/ns/:namespace", handlers.CreateNamespace(s.ns, s.pol))
	s.s.Get("/api/ns/:namespace", handlers.GetNamespace(s.ns, s.pol))
	s.s.Post("/api/ns/:namespace/r/:repo", handlers.CreateRepository(s.ns, s.rs, s.pol))
	s.s.Get("/api/ns/:namespace/r/:repo", handlers.GetRepository(s.ns, s.rs, s.pol))

	s.s.Get("/*", handlers.Repository)

}

// constructMiddleware adds all of the middleware required into the server, this is all of the defaults
func (s *Server) constructMiddleware() {
	s.s.Use(
		middleware.Cors,
		middleware.JsonHeaders,
		middleware.Authn(s.authn),
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
