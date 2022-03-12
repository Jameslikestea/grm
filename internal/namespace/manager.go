package namespace

import "github.com/Jameslikestea/grm/internal/models"

type Manager interface {
	CreateNamespace(req models.CreateNamespaceRequest) models.Namespace
	CreateNamespacePermission(ns models.NamespacePermission)
	GetNamespace(name string) (models.Namespace, error)
}
