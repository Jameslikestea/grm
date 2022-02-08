package git

import (
	"io"
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
)

type ReportItem struct {
	Ok     bool
	Reason string
}

type Report map[plumbing.ReferenceName]ReportItem

func (r Report) Write(w io.Writer) {
	e := pktline.NewEncoder(w)
	e.Encodef("unpack ok\n")
	keys := make([]string, len(r))
	i := 0
	for key, _ := range r {
		keys[i] = key.String()
		i++
	}
	sort.Strings(keys)
	for _, ref := range keys {
		item := r[plumbing.ReferenceName(ref)]
		if item.Ok {
			e.Encodef("ok %s\n", ref)
		} else {
			e.Encodef("ng %s %s\n", ref, item.Reason)
		}
	}
	e.Flush()
}
