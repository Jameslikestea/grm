package pubkey

import "github.com/Jameslikestea/grm/internal/models"

type Manager interface {
	StoreKey(pub models.UserPubKey) error
	GetKey(pub string) (models.UserPubKey, error)
	DeleteKey(pub string) error
}
