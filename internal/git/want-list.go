package git

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/rs/zerolog/log"
)

type WantList map[plumbing.Hash]bool
type HaveList map[plumbing.Hash]bool

func ParseWantList(wl WantList, reader io.Reader) (WantList, error) {
	e := pktline.NewScanner(reader)
	e.Scan()
	for {
		b := e.Bytes()
		if bytes.Equal(b, pktline.Flush) || bytes.Equal(b, []byte("done")) {
			return wl, nil
		}
		c := string(b)
		c = strings.TrimSpace(c)
		types := strings.Split(c, " ")
		if len(types) < 2 {
			return nil, errors.New("cannot decode want list")
		}
		if types[0] == "want" {
			hash := plumbing.NewHash(types[1])
			wl[hash] = true
			log.Trace().Str("hash", hash.String()).Msg("upload-pack want")
		}

		if ok := e.Scan(); !ok {
			log.Error().Msg("cannot decode want list")
		}
	}

	return wl, nil
}

func ParseHaveList(hl HaveList, reader io.Reader) (HaveList, error) {
	e := pktline.NewScanner(reader)
	e.Scan()
	for {
		b := e.Bytes()
		if bytes.Equal(b, pktline.Flush) || bytes.Equal(b, []byte("done\n")) {
			return hl, nil
		}
		log.Trace().Bytes("bytes", b).Msg("received want packet")
		c := string(b)
		c = strings.TrimSpace(c)
		types := strings.Split(c, " ")
		if len(types) < 2 {
			return nil, errors.New("cannot decode have list")
		}
		if types[0] == "hl" {
			hash := plumbing.NewHash(types[1])
			hl[hash] = true
			log.Trace().Str("hash", hash.String()).Msg("upload-pack have")
		}

		if ok := e.Scan(); !ok {
			log.Error().Msg("cannot decode want list")
		}
	}
}
