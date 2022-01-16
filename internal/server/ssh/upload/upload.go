package upload

import (
	"bytes"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHUploadPack(ch ssh.Channel, repo string, stor storage.Storage) {
	advertiseRefs(ch, stor, repo)
	wants, haves, _ := decodeRefs(ch, stor, repo)
	objs := getNewObjects(stor, repo, wants, haves)

	log.Trace().Int("objs", len(objs)).Msg("Counted number of objects")

}

func advertiseRefs(ch ssh.Channel, stor storage.Storage, repo string) {
	refs, err := stor.ListReferences(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot list references for advertise pack")
	}
	git.GenerateReferencePack(refs, false, "git-upload-pack", ch)
}

func decodeRefs(ch ssh.Channel, stor storage.Storage, repo string) (
	[]plumbing.Hash,
	map[plumbing.Hash]bool,
	map[string]bool,
) {
	wants := []plumbing.Hash{}
	haves := map[plumbing.Hash]bool{}
	capabilities := map[string]bool{}

	e := pktline.NewScanner(ch)
	e.Scan()
	for {
		b := e.Bytes()
		if bytes.Equal(b, pktline.Flush) {
			return wants, haves, capabilities
		}
		c := string(b)
		c = strings.TrimSpace(c)
		types := strings.Split(c, " ")
		if len(types) < 2 {
			continue
		}
		if len(types) > 2 {
			for _, cap := range types[2:] {
				capabilities[cap] = true
			}
		}

		h := plumbing.NewHash(types[1])

		switch types[0] {
		case "want":
			wants = append(wants, h)
		case "have":
			haves[h] = true
		default:
			continue
		}
		if ok := e.Scan(); !ok {
			log.Error().Err(e.Err()).Msg("done")
		}
	}
}

func getNewObjects(
	stor storage.Storage,
	repo string,
	wants []plumbing.Hash,
	haves map[plumbing.Hash]bool,
) []storage.Object {
	objs, err := stor.ListObjects(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get objects from store")
		return nil
	}
	cache := map[plumbing.Hash]storage.Object{}
	for _, obj := range objs {
		cache[obj.Hash] = obj
	}

	for _, want := range wants {
		c, ok := cache[want]
		log.Trace().Str("hash", want.String()).Bool("found", ok).Msg("Searching for objects")
		if !ok {
			continue
		}
		log.Trace().Str("type", c.Type.String()).Msg("Found hash")
	}

	return nil
}
