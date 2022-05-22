package service

import (
	"encoding/json"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/namespace"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ namespace.Manager = (*Service)(nil)

const hashSalt = "namespace:"
const hashPermSalt = "namespace:permission:"
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

func (s *Service) CreateNamespaceUserPermission(ns models.NamespacePermission) {
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+ns.Namespace+":"+ns.UserID))
	obj := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(ns),
	}
	s.stor.StoreObject(permRepo, obj, 0)
}

func (s *Service) UpdateNamespacePermissions(ns models.NamespacePermission) {
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+ns.Namespace))

	obj, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		// If there's an error then we assume that the object doesn't exist
		// we then create a permissions model for the whole thing.
		obj = storage.Object{
			Hash:    h,
			Type:    0,
			Content: models.Marshal([]models.NamespacePermission{ns}),
		}
		s.stor.StoreObject(permRepo, obj, 0)
		return
	}

	var permissions []models.NamespacePermission
	err = models.Unmarshal(obj.Content, &permissions)
	if err != nil {
		log.Warn().Str("hash", h.String()).Msg("Bad object, not a proper permission")
	}

	found := false
	for i, perm := range permissions {
		if perm.Namespace == ns.Namespace && perm.UserID == ns.UserID {
			found = true
			permissions[i] = ns
		}
	}

	if !found {
		permissions = append(permissions, ns)
	}

	obj = storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(permissions),
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

func (s *Service) ListNamespaces() ([]models.Namespace, error) {
	var ns []models.Namespace
	l, err := s.stor.ListObjects(repo)
	if err != nil {
		return ns, err
	}
	var n models.Namespace
	for _, o := range l {
		err = json.Unmarshal(o.Content, &n)
		if err != nil {
			return ns, err
		}
		ns = append(ns, n)
	}

	return ns, nil
}

func (s *Service) GetNamespacePermissions(namespace string) []models.NamespacePermission {
	var m []models.NamespacePermission
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+namespace))
	o, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		return m
	}
	models.Unmarshal(o.Content, &m)
	return m
}

func (s *Service) GetNamespaceUserPermissions(namespace, uid string) models.NamespacePermission {
	var m models.NamespacePermission
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+namespace+":"+uid))
	o, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		return m
	}
	models.Unmarshal(o.Content, &m)
	return m
}
