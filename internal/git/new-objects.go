package git

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/storage"
)

func GetNewObjects(
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
