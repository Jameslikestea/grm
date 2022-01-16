package git

import (
	"bytes"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/storage"
)

func TestGenerateReferencePack(t *testing.T) {
	type args struct {
		refs    []storage.Reference
		http    bool
		service string
	}
	tests := []struct {
		name       string
		args       args
		wantWriter string
	}{
		{
			name: "SSH No References Receive Pack",
			args: args{
				refs:    nil,
				http:    false,
				service: "git-receive-pack",
			},
			wantWriter: "00470000000000000000000000000000000000000000 capabilities^{}\x00ofs-delta\n0000",
		},
		{
			name: "HTTP No References Receive Pack",
			args: args{
				refs:    nil,
				http:    true,
				service: "git-receive-pack",
			},
			wantWriter: "001f# service=git-receive-pack\n00470000000000000000000000000000000000000000 capabilities^{}\x00ofs-delta\n0000",
		},
		{
			name: "SSH Single Reference",
			args: args{
				refs: []storage.Reference{
					{
						Name: "refs/tags/v1.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000043214321"),
					},
				},
				http:    false,
				service: "git-receive-pack",
			},
			wantWriter: "00480000000000000000000000000000000043214321 refs/tags/v1.0.0\x00ofs-delta\n0000",
		},
		{
			name: "HTTP Single Reference",
			args: args{
				refs: []storage.Reference{
					{
						Name: "refs/tags/v1.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000043214321"),
					},
				},
				http:    true,
				service: "git-receive-pack",
			},
			wantWriter: "001f# service=git-receive-pack\n00480000000000000000000000000000000043214321 refs/tags/v1.0.0\x00ofs-delta\n0000",
		},
		{
			name: "SSH Multi Reference",
			args: args{
				refs: []storage.Reference{
					{
						Name: "refs/tags/v1.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000043214321"),
					},
					{
						Name: "refs/tags/v2.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000012341234"),
					},
				},
				http:    false,
				service: "git-receive-pack",
			},
			wantWriter: "00480000000000000000000000000000000043214321 refs/tags/v1.0.0\x00ofs-delta\n003e0000000000000000000000000000000012341234 refs/tags/v2.0.0\n0000",
		},
		{
			name: "HTTP Multi Reference",
			args: args{
				refs: []storage.Reference{
					{
						Name: "refs/tags/v1.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000043214321"),
					},
					{
						Name: "refs/tags/v2.0.0",
						Hash: plumbing.NewHash("0000000000000000000000000000000012341234"),
					},
				},
				http:    true,
				service: "git-receive-pack",
			},
			wantWriter: "001f# service=git-receive-pack\n00480000000000000000000000000000000043214321 refs/tags/v1.0.0\x00ofs-delta\n003e0000000000000000000000000000000012341234 refs/tags/v2.0.0\n0000",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				writer := &bytes.Buffer{}
				GenerateReferencePack(tt.args.refs, tt.args.http, tt.args.service, writer)
				if gotWriter := writer.String(); gotWriter != tt.wantWriter {
					t.Errorf("GenerateReferencePack() = %v, want %v", gotWriter, tt.wantWriter)
				}
			},
		)
	}
}
