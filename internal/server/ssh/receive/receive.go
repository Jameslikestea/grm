package receive

import (
	"bytes"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/git"
	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHReceivePack(ch ssh.Channel, repo string, stor storage.Storage) {
	advertiseRefs(ch, stor, repo)
	refs, err := git.DecodeRefs(ch)
	if err != nil {
		ch.Stderr().Write([]byte("Cannot decode references\n"))
		return
	}
	objs := decodePack(ch, stor, repo)

	mapper := map[plumbing.Hash]storage.Object{}
	for _, obj := range objs {
		mapper[obj.Hash] = obj
	}

	stor.StoreObjects(repo, objs)
	stor.StoreReferences(repo, refs)
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
