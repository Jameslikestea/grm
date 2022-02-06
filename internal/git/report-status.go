package git

import (
	"io"

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
	for ref, item := range r {
		if item.Ok {
			e.Encodef("ok %s\n", ref.String())
		} else {
			e.Encodef("ng %s %s\n", ref.String(), item.Reason)
		}
	}
	e.Flush()
}
