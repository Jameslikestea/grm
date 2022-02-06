package git

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"

	"github.com/Jameslikestea/grm/internal/storage"
)

func GenerateReferencePack(refs []storage.Reference, http bool, service string, writer io.Writer) {
	e := pktline.NewEncoder(writer)
	if http {
		e.Encodef("# service=%s\n", service)
	}
	if len(refs) == 0 {
		e.Encodef(
			"%s %s\x00%s\n",
			plumbing.ZeroHash.String(),
			"capabilities^{}",
			"ofs-delta multi_ack report-status",
		)
	}
	for i, ref := range refs {
		if i == 0 {
			e.Encodef("%s %s\x00%s\n", ref.Hash.String(), ref.Name, "ofs-delta thin-pack multi_ack")
			continue
		}
		e.Encodef("%s %s\n", ref.Hash.String(), ref.Name)
	}

	e.Flush()
}
