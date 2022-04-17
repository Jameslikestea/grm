package policy

const (
	repo      = "data.repos"
	namespace = "data.namespaces"
)

const (
	RepoValidName   = repo + ".valid_name"
	RepoCreate      = repo + ".create"
	RepoRead        = repo + ".read"
	RepoWrite       = repo + ".write"
	RepoAdmin       = repo + ".admin"
	NamespaceCreate = namespace + ".create"
	NamespaceRead   = namespace + ".read"
	NamespaceWrite  = namespace + ".write"
	NamespaceAdmin  = namespace + ".admin"
)
