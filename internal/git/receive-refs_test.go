package git

import (
	"bytes"
	_ "embed"
	"io"
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/storage"
)

//go:embed content/receive-refs-valid
var Valid []byte

func TestDecodeRefs(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []storage.Reference
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				reader: bytes.NewReader(Valid),
			},
			want: []storage.Reference{
				{
					Name: "refs/heads/master",
					Hash: plumbing.NewHash("15027957951b64cf874c3557a0f3547bd83b3ff6"),
				},
				{
					Name: "refs/heads/experiment",
					Hash: plumbing.NewHash("cdfdb42577e2506715f8cfeacdbabc092bf63e8d"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := DecodeRefs(tt.args.reader)
				if (err != nil) != tt.wantErr {
					t.Errorf("DecodeRefs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DecodeRefs() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
