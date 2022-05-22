package namespace

import "github.com/Jameslikestea/grm/internal/models"

type Manager interface {
	CreateNamespace(req models.CreateNamespaceRequest) models.Namespace
	CreateNamespaceUserPermission(ns models.NamespacePermission)
	UpdateNamespacePermissions(ns models.NamespacePermission)

	GetNamespace(name string) (models.Namespace, error)
	GetNamespacePermissions(namespace string) []models.NamespacePermission
	GetNamespaceUserPermissions(namespace string, uid string) models.NamespacePermission
	ListNamespaces() ([]models.Namespace, error)
}
