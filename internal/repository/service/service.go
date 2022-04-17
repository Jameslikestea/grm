package service

import (
	"encoding/json"
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/repository"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ repository.Manager = (*Service)(nil)

const hashSalt = "repo:"
const hashPermSalt = "repo:permission:"
const hashNsSalt = "repo:ns:"
const repo = "_internal._repo"
const permRepo = "_internal._repo._permissions"
const nsRepo = "_internal._repo._namespace"

type Service struct {
	stor storage.Storage
}

func New(stor storage.Storage) *Service {
	return &Service{
		stor: stor,
	}
}

func (s *Service) CreateRepo(req models.CreateRepoRequest) models.Repo {
	re := models.Repo{
		Name:      req.Name,
		Namespace: req.Namespace,
		Public:    req.Public,
	}

	h := plumbing.ComputeHash(0, []byte(hashSalt+":"+re.Namespace+":"+re.Name))
	nsh := plumbing.ComputeHash(0, []byte(hashNsSalt+re.Namespace))

	repos, err := s.GetReposByNamespace(req.Namespace)
	if err != nil {
		repos = []models.Repo{
			re,
		}
	} else {
		repos = append(repos, re)
	}

	obj := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(re),
	}
	s.stor.StoreObject(repo, obj, 0)
	s.stor.StoreObject(nsRepo, storage.Object{Hash: nsh, Type: 0, Content: models.Marshal(repos)}, 0)

	return re
}

func (s *Service) CreateRepoUserPermission(ns models.RepoPermission) {
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+ns.RepoName+":"+ns.UserID))
	obj := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(ns),
	}
	s.stor.StoreObject(permRepo, obj, 0)
}

func (s *Service) UpdateRepoPermissions(ns models.RepoPermission) {
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+ns.RepoName))

	obj, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		// If there's an error then we assume that the object doesn't exist
		// we then create a permissions model for the whole thing.
		obj = storage.Object{
			Hash:    h,
			Type:    0,
			Content: models.Marshal([]models.RepoPermission{ns}),
		}
		s.stor.StoreObject(permRepo, obj, 0)
		return
	}

	var permissions []models.RepoPermission
	err = models.Unmarshal(obj.Content, &permissions)
	if err != nil {
		log.Warn().Str("hash", h.String()).Msg("Bad object, not a proper permission")
	}

	found := false
	for i, perm := range permissions {
		if perm.RepoName == ns.RepoName && perm.UserID == ns.UserID {
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

func (s *Service) GetRepo(namespace, r string) (models.Repo, error) {
	var ns models.Repo
	h := plumbing.ComputeHash(0, []byte(hashSalt+":"+namespace+":"+r))
	o, err := s.stor.GetObject(repo, h)
	if err != nil {
		return ns, err
	}

	err = json.Unmarshal(o.Content, &ns)

	return ns, err
}

func (s *Service) GetReposByNamespace(namespace string) ([]models.Repo, error) {
	var ns []models.Repo
	h := plumbing.ComputeHash(0, []byte(hashNsSalt+namespace))
	o, err := s.stor.GetObject(nsRepo, h)
	if err != nil {
		return ns, err
	}

	err = models.Unmarshal(o.Content, &ns)

	return ns, nil
}

func (s *Service) GetRepoPermissions(namespace, r string) []models.RepoPermission {
	var m []models.RepoPermission
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+r))
	o, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		return m
	}
	models.Unmarshal(o.Content, &m)
	return m
}

func (s *Service) GetRepoUserPermissions(namespace, r, uid string) models.RepoPermission {
	var m models.RepoPermission
	h := plumbing.ComputeHash(0, []byte(hashPermSalt+r+":"+uid))
	o, err := s.stor.GetObject(permRepo, h)
	if err != nil {
		return m
	}
	models.Unmarshal(o.Content, &m)
	return m
}

func (s *Service) GetTags(ns, r string) []string {
	var tags []string

	references, _ := s.stor.ListReferences(ns + "/" + r + ".git")

	for _, ref := range references {
		tags = append(tags, ref.Name.Short())
	}

	sort.Strings(tags)

	return tags
}
