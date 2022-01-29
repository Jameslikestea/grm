package git

import (
	"bytes"
	_ "embed"
	"io"
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

//go:embed content/valid_want_list
var validWantList []byte

//go:embed content/invalid_want_list_missing_types
var invalidWantListMissingHash []byte

func TestParseWantList(t *testing.T) {
	type args struct {
		wl     WantList
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    WantList
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				wl:     WantList{},
				reader: bytes.NewReader(validWantList),
			},
			want: WantList{
				plumbing.NewHash("74730d410fcb6603ace96f1dc55ea6196122532d"): true,
				plumbing.NewHash("7d1665144a3a975c05f1f43902ddaf084e784dbe"): true,
			},
		},
		{
			name: "Missing Hash",
			args: args{
				wl:     WantList{},
				reader: bytes.NewReader(invalidWantListMissingHash),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ParseWantList(tt.args.wl, tt.args.reader)
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseWantList() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseWantList() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
