package http

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/server/http/handlers"
	"github.com/Jameslikestea/grm/internal/server/http/middleware"
)

type Server struct {
	s *fiber.App
}

func NewServer() *Server {
	engine := html.New("./templates", ".html")

	s := &Server{
		s: fiber.New(
			fiber.Config{
				Views:                 engine,
				AppName:               "grmpkg",
				DisableStartupMessage: true,
			},
		),
	}

	s.constructMiddleware()
	s.constructRoutes()

	return s
}

// constructRoutes adds in all of the specific and generic route handlers
func (s *Server) constructRoutes() {
	s.s.Get("/", handlers.Index)
	s.s.Get("/package", handlers.Package)
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
