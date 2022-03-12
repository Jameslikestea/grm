package service

import (
	"encoding/json"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ namespace.Manager = (*Service)(nil)

const hashSalt = "namespace:"
const repo = "_internal._namespace"
const permRepo = "_internal._namespace._permissions"

type Service struct {
	stor storage.Storage
}

func New(stor storage.Storage) *Service {
	return &Service{
		stor: stor,
	}
}

func (s *Service) CreateNamespace(req models.CreateNamespaceRequest) models.Namespace {
	ns := models.Namespace{
		Name:   req.Name,
		Public: req.Public,
	}

	h := plumbing.ComputeHash(0, []byte(hashSalt+ns.Name))

	obj := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(ns),
	}
	s.stor.StoreObject(repo, obj, 0)

	return ns
}

func (s *Service) CreateNamespacePermission(ns models.NamespacePermission) {
	h := plumbing.ComputeHash(0, []byte(hashSalt+ns.Namespace+":"+ns.UserID))
	obj := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(ns),
	}
	s.stor.StoreObject(permRepo, obj, 0)
}

func (s *Service) GetNamespace(name string) (models.Namespace, error) {
	var ns models.Namespace
	h := plumbing.ComputeHash(0, []byte(hashSalt+name))
	o, err := s.stor.GetObject(repo, h)
	if err != nil {
		return ns, err
	}

	err = json.Unmarshal(o.Content, &ns)

	return ns, err
}
