package service

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/pubkey"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ pubkey.Manager = &Service{}

const hashSalt = "namespace:"
const repo = "_internal._sshkeys"

type Service struct {
	stor storage.Storage
}

func New(s storage.Storage) *Service {
	return &Service{
		stor: s,
	}
}

func (s *Service) normalizeKey(k string) (string, error) {
	pkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k))
	if err != nil {
		log.Error().Err(err).Msg("Cannot Parse Provided Key")
		return "", err
	}

	keyString := ssh.MarshalAuthorizedKey(pkey)
	return string(keyString), nil

}

func (s *Service) hash(k string) plumbing.Hash {
	h := plumbing.ComputeHash(0, []byte(hashSalt+k))
	return h
}

func (s *Service) StoreKey(pub models.UserPubKey) error {
	key, err := s.normalizeKey(pub.Key)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot normalize key")
		return err
	}
	log.Info().Str("uid", pub.UID).Msg("Storing Key for user")

	ukey, _ := s.GetKey(key)
	if ukey.Key == key {
		return errors.New("key already in use")
	}

	o := storage.Object{
		Hash: s.hash(key),
		Type: 0,
		Content: models.Marshal(
			models.UserPubKey{
				UID: pub.UID,
				Key: key,
			},
		),
	}

	err = s.stor.StoreObject(repo, o, 0)

	return err
}

func (s *Service) GetKey(pub string) (models.UserPubKey, error) {
	var pubKey models.UserPubKey

	key, err := s.normalizeKey(pub)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot normalize key")
		return pubKey, err
	}

	object, err := s.stor.GetObject(repo, s.hash(key))
	if err != nil {
		return pubKey, nil
	}

	err = models.Unmarshal(object.Content, &pubKey)
	if err != nil {
		return pubKey, err
	}

	return pubKey, nil
}

func (s *Service) DeleteKey(pub string) error {
	key, err := s.normalizeKey(pub)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot normalize key")
		return err
	}

	err = s.stor.StoreObject(
		repo, storage.Object{
			Hash:    s.hash(key),
			Type:    0,
			Content: []byte{},
		}, 1,
	)

	return err
}
