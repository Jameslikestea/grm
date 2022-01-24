package upload

import (
	"bytes"
	"io"
	"strings"

	packfile2 "gg-scm.io/pkg/git/packfile"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHUploadPack(ch ssh.Channel, repo string, stor storage.Storage) {
	advertiseRefs(ch, stor, repo)
	wants, haves, _ := decodeRefs(ch, stor, repo)

	log.Info().Int("haves", len(haves)).Msg("Received Haves")
	objs := getNewObjects(stor, repo, wants, haves)
	ch.Write([]byte("0008NAK\n"))
	encode(ch, objs)

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

	flush_count := 0
	for {
		b := e.Bytes()
		log.Trace().Bytes("content", b).Msg("Received pktline")
		if bytes.Equal(b, pktline.Flush) || bytes.Equal(b, []byte("done\n")) {
			flush_count++
			log.Trace().Msg("Received pktline.Flush")
			if flush_count == 2 {
				return wants, haves, capabilities
			}
			e = pktline.NewScanner(ch)
			e.Scan()
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
			log.Trace().Str("hash", h.String()).Msg("Got Want")
			wants = append(wants, h)
		case "have":
			log.Trace().Str("hash", h.String()).Msg("Got Have")
			haves[h] = true
		default:
			log.Trace().Str("hash", h.String()).Str("type", types[0]).Msg("Defaulting")
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
	seen := map[plumbing.Hash]bool{}

	for _, obj := range objs {
		cache[obj.Hash] = obj
	}

	log.Info().Int("haves", len(haves)).Msg("Haves")
	for key := range haves {
		recurseFound(cache, seen, key)
	}

	log.Info().Int("seen", len(seen)).Msg("Found Seen")

	newObjs := map[plumbing.Hash]storage.Object{}
	for _, want := range wants {
		recurseFind(cache, newObjs, seen, want)
	}

	log.Info().Int("discovered", len(newObjs)).Msg("Found New")
	objList := make([]storage.Object, len(newObjs))
	i := 0
	for _, obj := range newObjs {
		objList[i] = obj
		i++
	}

	return objList
}

func recurseFound(cache map[plumbing.Hash]storage.Object, seen map[plumbing.Hash]bool, hash plumbing.Hash) {
	// Do nothing if we've already seen the hash, prevents circular searching
	if _, ok := seen[hash]; ok {
		log.Trace().Str("hash", hash.String()).Msg("Already seen hash")
		return
	}
	seen[hash] = true
	obj, ok := cache[hash]
	if !ok {
		// We don't need to throw an error as this will be handled elsewhere
		return
	}
	m := &plumbing.MemoryObject{}
	switch obj.Type {
	case plumbing.CommitObject:
		log.Trace().Msg("Searching Commit")
		m.SetType(obj.Type)
		m.SetSize(int64(len(obj.Content)))
		m.Write(obj.Content)

		c := object.Commit{}
		c.Decode(m)
		recurseFound(cache, seen, c.TreeHash)
		for _, ph := range c.ParentHashes {
			recurseFound(cache, seen, ph)
		}

	case plumbing.TreeObject:
		log.Trace().Msg("Searching Tree")
		m.SetType(obj.Type)
		m.SetSize(int64(len(obj.Content)))
		m.Write(obj.Content)

		c := object.Tree{}
		c.Decode(m)

		for _, entry := range c.Entries {
			recurseFound(cache, seen, entry.Hash)
		}
	}

}

func recurseFind(
	cache, objs map[plumbing.Hash]storage.Object,
	seen map[plumbing.Hash]bool,
	hash plumbing.Hash,
) {
	if _, ok := seen[hash]; ok {
		return
	}
	obj, ok := cache[hash]
	if !ok {
		return
	}
	// Setup the object and mark as now seen
	objs[hash] = obj
	seen[hash] = true

	m := &plumbing.MemoryObject{}
	switch obj.Type {
	case plumbing.CommitObject:
		m.SetType(obj.Type)
		m.SetSize(int64(len(obj.Content)))
		m.Write(obj.Content)

		c := object.Commit{}
		c.Decode(m)
		recurseFind(cache, objs, seen, c.TreeHash)
		for _, ph := range c.ParentHashes {
			recurseFind(cache, objs, seen, ph)
		}

	case plumbing.TreeObject:
		m.SetType(obj.Type)
		m.SetSize(int64(len(obj.Content)))
		m.Write(obj.Content)

		c := object.Tree{}
		c.Decode(m)

		for _, entry := range c.Entries {
			recurseFind(cache, objs, seen, entry.Hash)
		}
	}

}

func encode(w io.Writer, objs []storage.Object) {
	writer := packfile2.NewWriter(w, uint32(len(objs)))

	for _, obj := range objs {
		hdr := &packfile2.Header{
			Type: packfile2.ObjectType(obj.Type),
			Size: int64(len(obj.Content)),
		}

		writer.WriteHeader(hdr)
		writer.Write(obj.Content)
	}

	writer.Close()
}
