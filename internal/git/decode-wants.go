package git

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"

	"github.com/Jameslikestea/grm/internal/storage"
)

func DecodeWants(ch io.Reader, stor storage.Storage, repo string) (
	[]plumbing.Hash,
	map[plumbing.Hash]bool,
	map[string]bool,
) {
	wants := []plumbing.Hash{}
	var haves map[plumbing.Hash]bool

	e := pktline.NewScanner(ch)
	e.Scan()

	w := WantList{}
	h := HaveList{}

	ParseWantList(w, ch)
	ParseHaveList(h, ch)

	haves = h
	for k := range w {
		wants = append(wants, k)
	}

	return wants, haves, map[string]bool{}
}
