package receive

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"

	"github.com/Jameslikestea/grm/internal/storage"
)

func SSHReceivePack(ch ssh.Channel, repo string, stor storage.Storage) {
	advertiseRefs(ch, stor, repo)
	refs := decodeRefs(ch)
	objs := decodePack(ch, stor, repo)

	stor.StoreObjects(repo, objs)
	stor.StoreReferences(repo, refs)
}

func advertiseRefs(ch ssh.Channel, stor storage.Storage, repo string) {
	refs, _ := stor.ListReferences(repo)
	e := pktline.NewEncoder(ch)
	if len(refs) == 0 {
		e.Encodef("%s %s\x00%s\n", plumbing.ZeroHash.String(), "capabilities^{}", "ofs-delta")
		e.Flush()
		return
	}

	for i, ref := range refs {
		if i == 0 {
			e.Encodef("%s %s\x00%s\n", ref.Hash.String(), ref.Name, "ofs-delta")
		} else {
			e.Encodef("%s %s\n", ref.Hash.String(), ref.Name)
		}
	}

	e.Flush()
}

func decodeRefs(ch ssh.Channel) []storage.Reference {
	rfs := []storage.Reference{}
	e := pktline.NewScanner(ch)
	e.Scan()
	for {
		b := e.Bytes()
		if bytes.Equal(b, pktline.Flush) {
			log.Info().Msg("Received end packet")
			return rfs
		}

		src := plumbing.NewHash(string(b[0:40]))
		dst := plumbing.NewHash(string(b[41:81]))
		ref := plumbing.ReferenceName(
			strings.Map(
				func(r rune) rune {
					if unicode.IsControl(r) {
						return -1
					}
					return r
				}, string(b[82:]),
			),
		)

		rfs = append(rfs, storage.Reference{Name: ref, Hash: dst})

		log.Debug().Str("src", src.String()).Str("dst", dst.String()).Str(
			"ref",
			ref.Short(),
		).Bool("tag", ref.IsTag()).Msg("Received reference update request")

		if ok := e.Scan(); !ok {
			log.Error().Err(e.Err()).Msg("done")
		}
	}
	return rfs
}

func decodePack(ch ssh.Channel, stor storage.Storage, repo string) []storage.Object {
	objs := []storage.Object{}

	bufs := map[int64][]byte{}
	hahs := map[plumbing.Hash][]byte{}

	btyp := map[int64]plumbing.ObjectType{}
	htyp := map[plumbing.Hash]plumbing.ObjectType{}

	obs, _ := stor.ListObjects(repo)
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
