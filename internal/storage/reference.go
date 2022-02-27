package storage

import (
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
)

type Reference struct {
	Name plumbing.ReferenceName
	Hash plumbing.Hash
}

type Object struct {
	Hash    plumbing.Hash
	Type    plumbing.ObjectType
	Content []byte
}

type HashKey struct {
	KID gocql.UUID
	K   string
}

type AuthenticationSession struct {
	User      string     `json:"uid"`
	TID       gocql.UUID `json:"tid"`
	KID       gocql.UUID `json:"kid"`
	Type      string     `json:"type"`
	Signature string     `json:"signature"`
}

func (a AuthenticationSession) String() string {
	s, _ := json.Marshal(a)
	return string(s)
}

func (a AuthenticationSession) UnhashedString() string {
	return fmt.Sprintf("%s:%s:%s", a.TID, a.Type, a.User)
}

type Storage interface {
	StoreReferences(string, []Reference) error
	StoreObjects(string, []Object) error
	StoreObject(string, Object, int) error
	ListReferences(string) ([]Reference, error)
	ListObjects(string) ([]Object, error)
	GetObject(string, plumbing.Hash) (Object, error)

	GenerateHashKey() error
	GetHashKey() ([]HashKey, error)
}
