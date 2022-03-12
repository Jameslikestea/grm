package models

type Namespace struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type NamespacePermission struct {
	UserID    string `json:"uid"`
	Namespace string `json:"namespace"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Admin     bool   `json:"admin"`
}

type CreateNamespaceRequest struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}
