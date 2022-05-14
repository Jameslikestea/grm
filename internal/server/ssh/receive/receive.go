package receive

import (
	"bytes"
	"fmt"

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

		for _, ref := range validRefs {
			ch.Stderr().Write([]byte(fmt.Sprintf("Server Validating: %s\n", ref.Name)))
			// if !validateRef(ref, mapper) {
			// 	report[ref.Name] = git.ReportItem{
			// 		Ok:     false,
			// 		Reason: "Could not validate object tree",
			// 	}
			// }
		}

		stor.StoreObjects(repo, objs)
		stor.StoreReferences(repo, validRefs)
	}

	report.Write(ch)
}

func validateRef(reference storage.Reference, objs map[plumbing.Hash]storage.Object) bool {
	obj, ok := objs[reference.Hash]

	if !ok {
		return false
	}

	return validateObj(obj, objs)
}

func validateObj(obj storage.Object, objs map[plumbing.Hash]storage.Object) bool {
	switch obj.Type {
	case plumbing.CommitObject:
		o := &plumbing.MemoryObject{}
		o.SetType(plumbing.CommitObject)
		o.SetSize(int64(len(obj.Content)))
		o.Write(obj.Content)
		commit := object.Commit{}
		err := commit.Decode(o)
		if err != nil {
			return false
		}
		log.Debug().Msg("validating commit tree")
		tree, ok := objs[commit.TreeHash]
		if !ok {
			return false
		}
		if !validateObj(tree, objs) {
			return false
		}

		log.Debug().Int("parents", len(commit.ParentHashes)).Msg("validating commit parents")
		for _, parentHash := range commit.ParentHashes {
			parent, ok := objs[parentHash]
			if !ok {
				return false
			}
			if !validateObj(parent, objs) {
				return false
			}
		}

		return true
	case plumbing.TreeObject:
		o := &plumbing.MemoryObject{}
		o.SetType(plumbing.TreeObject)
		o.SetSize(int64(len(obj.Content)))
		o.Write(obj.Content)
		tree := object.Tree{}
		err := tree.Decode(o)
		if err != nil {
			return false
		}

		log.Debug().Int("entries", len(tree.Entries)).Msg("validating tree entries")
		for _, entry := range tree.Entries {
			ob, ok := objs[entry.Hash]
			if !ok {
				return false
			}
			if !validateObj(ob, objs) {
				return false
			}
		}
		return true
	case plumbing.BlobObject:
		log.Debug().Msg("validating blob")
		return true
	case plumbing.TagObject:
		return true
	case plumbing.OFSDeltaObject:
		return true
	case plumbing.REFDeltaObject:
		return true
	}
	return true
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

			m := &plumbing.MemoryObject{}
			m.Write(bts)
			m.SetType(t)
			hash := m.Hash()

			bufs[header.Offset] = bts
			btyp[header.Offset] = t
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

			m := &plumbing.MemoryObject{}
			m.Write(bts)
			m.SetType(t)
			hash := m.Hash()

			bufs[header.Offset] = bts
			btyp[header.Offset] = t
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
