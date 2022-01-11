package memory

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/storage"
)

type MemoryStorage struct {
	refs    map[string]plumbing.Hash
	objects map[plumbing.Hash]storage.Object
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		refs:    make(map[string]plumbing.Hash),
		objects: make(map[plumbing.Hash]storage.Object),
	}
}

func (m MemoryStorage) StoreReferences(repo string, references []storage.Reference) error {
	for _, ref := range references {
		if _, ok := m.objects[ref.Hash]; !ok {
			continue
		}
		m.refs[ref.Name.String()] = ref.Hash
	}
	return nil
}

func (m MemoryStorage) StoreObjects(repo string, objects []storage.Object) error {
	for _, obj := range objects {
		m.objects[obj.Hash] = obj
	}
	return nil
}

func (m MemoryStorage) GetObject(repo string, hash plumbing.Hash) (storage.Object, error) {
	obj, ok := m.objects[hash]
	if !ok {
		return storage.Object{}, errors.New("no such object")
	}
	return obj, nil

}

func (m MemoryStorage) ListReferences(repo string) ([]storage.Reference, error) {
	l := []storage.Reference{}
	for r, h := range m.refs {
		l = append(l, storage.Reference{Name: plumbing.ReferenceName(r), Hash: h})
	}

	return l, nil
}

func (m MemoryStorage) ListObjects(repo string) ([]storage.Object, error) {
	l := []storage.Object{}
	for _, h := range m.objects {
		l = append(l, h)
	}

	return l, nil
}
