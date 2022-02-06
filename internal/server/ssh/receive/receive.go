package receive

import (
	"bytes"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHReceivePack(ch ssh.Channel, repo string, stor storage.Storage) {
	// Stage 1 is to receive information from the client.
	advertiseRefs(ch, stor, repo)
	refs, err := git.DecodeRefs(ch)
	if err != nil {
		ch.Stderr().Write([]byte("Cannot decode references\n"))
		return
	}
	objs := decodePack(ch, stor, repo)

	report := git.Report{}
	validRefs := []storage.Reference{}
	for _, ref := range refs {
		if ref.Name.IsTag() {
			validRefs = append(validRefs, ref)
		}
		report[ref.Name] = git.ReportItem{
			Ok:     ref.Name.IsTag(),
			Reason: "GRM only accepts tags",
		}
	}

	validateTags(report, repo, stor)

	if len(validRefs) > 0 {

		mapper := map[plumbing.Hash]storage.Object{}
		for _, obj := range objs {
			mapper[obj.Hash] = obj
		}

		stor.StoreObjects(repo, objs)
		stor.StoreReferences(repo, validRefs)
	}

	report.Write(ch)
}

// validateObjects validates that the objects are required for the repository to function
func validateObjects(
	validRefs []storage.Reference,
	repoObjects []storage.Object,
	repo string,
	stor storage.Storage,
) []storage.Object {
	currentObjs, err := stor.ListObjects(repo)
	if err != nil {
		log.Warn().Err(err).Msg("failed to validate objects")
		return repoObjects
	}

	seen := map[plumbing.Hash]bool{}
	nObjs := map[plumbing.Hash]storage.Object{}
	cache := map[plumbing.Hash]storage.Object{}

	for _, obj := range currentObjs {
		seen[obj.Hash] = true
	}

	for _, obj := range repoObjects {
		cache[obj.Hash] = obj
	}

	for _, ref := range validRefs {
		hunt(ref.Hash, cache, nObjs, seen)
	}

	newObjects := []storage.Object{}
	for _, obj := range nObjs {
		if _, ok := seen[obj.Hash]; !ok {
			newObjects = append(newObjects, obj)
		}
	}

	return newObjects
}

func hunt(hash plumbing.Hash, cache, nObjs map[plumbing.Hash]storage.Object, seen map[plumbing.Hash]bool) {
	if _, ok := seen[hash]; ok {
		return
	}
	obj, ok := cache[hash]
	if !ok {
		log.Warn().Msg("Cannot find object in cache")
		return
	}

	nObjs[hash] = obj
	seen[hash] = true

	m := &plumbing.MemoryObject{}
	m.SetType(obj.Type)
	m.SetSize(int64(len(obj.Content)))
	m.Write(obj.Content)

	switch obj.Type {
	case plumbing.CommitObject:
		c := object.Commit{}
		c.Decode(m)
		for _, h := range c.ParentHashes {
			hunt(h, cache, nObjs, seen)
		}
	case plumbing.TreeObject:
		t := object.Tree{}
		t.Decode(m)
		for _, h := range t.Entries {
			hunt(h.Hash, cache, nObjs, seen)
		}
	}
}

func validateTags(r git.Report, repo string, stor storage.Storage) {
	ref, err := stor.ListReferences(repo)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot validate tags")
		return
	}

	for _, rf := range ref {
		if reference, ok := r[rf.Name]; ok {
			reference.Ok = false
			reference.Reason = "GRM tags are immutable"
			r[rf.Name] = reference
		}
	}
}

func advertiseRefs(ch ssh.Channel, stor storage.Storage, repo string) {
	refs, err := stor.ListReferences(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot list references for advertise pack")
	}
	git.GenerateReferencePack(refs, false, "git-receive-pack", ch)
}

func decodePack(ch ssh.Channel, stor storage.Storage, repo string) []storage.Object {
	objs := []storage.Object{}

	bufs := map[int64][]byte{}
	hahs := map[plumbing.Hash][]byte{}

	btyp := map[int64]plumbing.ObjectType{}
	htyp := map[plumbing.Hash]plumbing.ObjectType{}

	obs, err := stor.ListObjects(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get objects from DB")
	}
	for _, obj := range obs {
		hahs[obj.Hash] = obj.Content
		htyp[obj.Hash] = obj.Type
	}

	reader := packfile.NewScanner(ch)

	v, o, err := reader.Header()
	if err != nil {
		log.Error().Err(err).Msg("Couldn't decode packfile")
	}
	log.Info().Uint32("version", v).Uint32("objects", o).Msg("Receiving Packfiles")

	for i := uint32(0); i < o; i++ {
		b := bytes.NewBuffer([]byte(""))
		header, _ := reader.NextObjectHeader()

		reader.NextObject(b)

		switch header.Type {
		case plumbing.CommitObject, plumbing.TreeObject, plumbing.BlobObject, plumbing.TagObject:
			bts := b.Bytes()
			bufs[header.Offset] = bts
			btyp[header.Offset] = header.Type
			hash := plumbing.ComputeHash(header.Type, bts)
			hahs[hash] = bts
			htyp[hash] = header.Type

			log.Trace().Int64("offset", header.Offset).Int("length", b.Len()).Str(
				"type",
				header.Type.String(),
			).Msg("received object")

		case plumbing.OFSDeltaObject:
			log.Trace().Int64("offset", header.Offset).Int64(
				"offset-reference",
				header.OffsetReference,
			).Msg("received ofs delta")

			src, ok := bufs[header.OffsetReference]
			if !ok {
				log.Error().Msg("Cannot find offset in pack")
			}
			t := btyp[header.OffsetReference]

			s, d := src, b.Bytes()

			bts, err := packfile.PatchDelta(s, d)
			if err != nil {
				log.Error().Err(err).Bytes("src", s).Bytes("delta", d).Msg("Cannot apply delta")
			}
			bufs[header.Offset] = bts
			btyp[header.Offset] = t
			hash := plumbing.ComputeHash(t, bts)
			hahs[hash] = bts
			htyp[hash] = t

		case plumbing.REFDeltaObject:
			log.Trace().Str("reference", header.Reference.String()).Msg("received ref delta")

			src, ok := hahs[header.Reference]
			if !ok {
				log.Error().Msg("Cannot find offset in pack")
			}
			t := htyp[header.Reference]
			s, d := src, b.Bytes()

			bts, err := packfile.PatchDelta(s, d)
			if err != nil {
				log.Error().Err(err).Bytes("src", s).Bytes("delta", d).Msg("Cannot apply delta")
			}
			bufs[header.Offset] = bts
			btyp[header.Offset] = t
			hash := plumbing.ComputeHash(t, bts)
			hahs[hash] = bts
			htyp[hash] = t
		}

		b.Reset()
	}

	for hash, content := range hahs {
		objs = append(
			objs, storage.Object{
				Hash:    hash,
				Type:    htyp[hash],
				Content: content,
			},
		)
	}

	return objs
}
