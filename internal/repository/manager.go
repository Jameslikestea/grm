package repository

import "github.com/Jameslikestea/grm/internal/models"

type Manager interface {
	CreateRepo(req models.CreateRepoRequest) models.Repo
	CreateRepoUserPermission(ns models.RepoPermission)
	UpdateRepoPermissions(ns models.RepoPermission)

	GetRepo(namespace, repo string) (models.Repo, error)
	GetRepoPermissions(namespace, repo string) []models.RepoPermission
	GetRepoUserPermissions(namespace, repo, uid string) models.RepoPermission
}
