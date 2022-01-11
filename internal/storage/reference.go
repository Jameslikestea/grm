package storage

import "github.com/go-git/go-git/v5/plumbing"

type Reference struct {
	Name plumbing.ReferenceName
	Hash plumbing.Hash
}

type Object struct {
	Hash    plumbing.Hash
	Type    plumbing.ObjectType
	Content []byte
}

type Storage interface {
	StoreReferences(string, []Reference) error
	StoreObjects(string, []Object) error
	ListReferences(string) ([]Reference, error)
	ListObjects(string) ([]Object, error)
	GetObject(string, plumbing.Hash) (Object, error)
}
