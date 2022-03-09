package policy

type PolicyRequest struct {
	Repo                 Repo                  `json:"repo"`
	Namespace            Namespace             `json:"namespace"`
	RepoPermissions      []RepoPermission      `json:"repo_permissions"`
	NamespacePermissions []NamespacePermission `json:"namespace_permissions"`
	UserID               string                `json:"uid"`
}

type Repo struct {
	Name      string    `json:"name"`
	Namespace Namespace `json:"namespace"`
	Public    bool      `json:"public"`
}

type Namespace struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type RepoPermission struct {
	UserID   string `json:"uid"`
	RepoName string `json:"repo_name"`
	Read     bool   `json:"read"`
	Write    bool   `json:"write"`
	Admin    bool   `json:"admin"`
}

type NamespacePermission struct {
	UserID    string `json:"uid"`
	Namespace string `json:"namespace"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Admin     bool   `json:"admin"`
}
