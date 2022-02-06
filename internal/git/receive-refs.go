package git

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/storage"
)

func DecodeRefs(reader io.Reader) ([]storage.Reference, error) {
	var pl []storage.Reference

	e := pktline.NewScanner(reader)
	if ok := e.Scan(); !ok {
		return nil, errors.New("cannot read pktline")
	}
	for {
		b := e.Bytes()
		if bytes.Equal(b, pktline.Flush) {
			log.Info().Msg("Received end packet")
			return pl, nil
		}

		components := strings.Split(string(b), " ")
		if len(components) < 3 {
			return nil, errors.New("cannot read pktline")
		}

		src := plumbing.NewHash(components[0])
		dst := plumbing.NewHash(components[1])
		ref := plumbing.ReferenceName(
			strings.Map(
				func(r rune) rune {
					if unicode.IsControl(r) {
						return -1
					}
					return r
				}, components[2],
			),
		)

		pl = append(pl, storage.Reference{Name: ref, Hash: dst})

		log.Debug().Str("src", src.String()).Str("dst", dst.String()).Str(
			"ref",
			ref.Short(),
		).Bool("tag", ref.IsTag()).Msg("Received reference update request")

		if ok := e.Scan(); !ok {
			log.Error().Err(e.Err()).Msg("done")
		}
	}
}
