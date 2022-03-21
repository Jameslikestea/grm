package policy

import (
	"context"
	"embed"
	"io/fs"
	"io/ioutil"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/rs/zerolog/log"
)

var _ Manager = (*Service)(nil)

type Service struct {
	r *ast.Compiler
}

//go:embed policies/**/*.rego
var policies embed.FS

func New() *Service {
	r := map[string]string{}
	fs.WalkDir(
		policies, ".", func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				log.Info().Str("path", path).Msg("Loading Policy")
				f, _ := policies.Open(path)
				b, _ := ioutil.ReadAll(f)
				r[path] = string(b)
			}
			return nil
		},
	)

	re, err := ast.CompileModules(r)
	if err != nil {
		log.Error().Err(err).Msg("Cannot compile modules")
		return nil
	}

	return &Service{
		r: re,
	}
}

func (s *Service) Evaluate(query string, input interface{}) bool {
	log.Info().Interface("input", input).Msg("Querying Policies")
	rs, err := rego.New(
		rego.Query(query), rego.Compiler(s.r), rego.Input(input),
	).Eval(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to evaluate query")
		return false
	}

	return rs.Allowed()
}
