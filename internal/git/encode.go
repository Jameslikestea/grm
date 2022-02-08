package git

import (
	"io"

	"gg-scm.io/pkg/git/packfile"

	"github.com/Jameslikestea/grm/internal/storage"
)

func EncodePackfile(w io.Writer, objs []storage.Object) {
	writer := packfile.NewWriter(w, uint32(len(objs)))

	for _, obj := range objs {
		hdr := &packfile.Header{
			Type: packfile.ObjectType(obj.Type),
			Size: int64(len(obj.Content)),
		}

		writer.WriteHeader(hdr)
		writer.Write(obj.Content)
	}

	writer.Close()
}
