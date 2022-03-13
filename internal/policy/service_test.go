package policy

import (
	"testing"

	"github.com/Jameslikestea/grm/internal/models"
)

func TestService_Evaluate(t *testing.T) {
	type args struct {
		query string
		input PolicyRequest
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid named repo",
			args: args{
				query: RepoValidName,
				input: PolicyRequest{
					Repo: models.Repo{
						Name: "repository",
					},
				},
			},
			want: true,
		},
		{
			name: "Invalid named repo",
			args: args{
				query: RepoValidName,
				input: PolicyRequest{
					Repo: models.Repo{
						Name: "_internal.myrepo",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := New()
				if got := s.Evaluate(tt.args.query, tt.args.input); got != tt.want {
					t.Errorf("Evaluate() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
