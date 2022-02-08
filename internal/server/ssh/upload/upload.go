package upload

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHUploadPack(ch ssh.Channel, repo string, stor storage.Storage) {
	advertiseRefs(ch, stor, repo)
	wants, haves, _ := git.DecodeWants(ch, stor, repo)

	log.Info().Int("haves", len(haves)).Msg("Received Haves")
	objs := git.GetNewObjects(stor, repo, wants, haves)
	ch.Write([]byte("0008NAK\n"))
	git.EncodePackfile(ch, objs)

	log.Trace().Int("objs", len(objs)).Msg("Counted number of objects")

}

func advertiseRefs(ch ssh.Channel, stor storage.Storage, repo string) {
	refs, err := stor.ListReferences(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot list references for advertise pack")
	}
	git.GenerateReferencePack(refs, false, "git-upload-pack", ch)
}
