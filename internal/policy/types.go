package policy

import "github.com/Jameslikestea/grm/internal/models"

type PolicyRequest struct {
	Repo                 models.Repo                   `json:"repo"`
	CreateNamespace      models.CreateNamespaceRequest `json:"create_namespace"`
	Namespace            models.Namespace              `json:"namespace"`
	CreateRepo           models.CreateRepoRequest      `json:"create_repo"`
	RepoPermissions      []models.RepoPermission       `json:"repo_permissions"`
	NamespacePermissions []models.NamespacePermission  `json:"namespace_permissions"`
	UserID               string                        `json:"uid"`
}
