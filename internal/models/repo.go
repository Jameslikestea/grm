package models

var _ Model = Repo{}
var _ Model = RepoPermission{}
var _ Model = CreateRepoRequest{}

type Repo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Public    bool   `json:"public"`
}

type RepoPermission struct {
	UserID   string `json:"uid"`
	RepoName string `json:"repo_name"`
	Read     bool   `json:"read"`
	Write    bool   `json:"write"`
	Admin    bool   `json:"admin"`
}

type CreateRepoRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Public    bool   `json:"public"`
}
