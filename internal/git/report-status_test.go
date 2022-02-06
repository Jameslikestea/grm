package git

import (
	"bytes"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestReport_Write(t *testing.T) {
	tests := []struct {
		name  string
		r     Report
		wantW string
	}{
		{
			name: "Single OK",
			r: Report{
				plumbing.ReferenceName("refs/heads/master"): ReportItem{
					Ok:     true,
					Reason: "",
				},
			},
			wantW: "000eunpack ok\n0019ok refs/heads/master\n0000",
		},
		{
			name: "Tags vs Branches",
			r: Report{
				plumbing.ReferenceName("refs/heads/master"): ReportItem{
					Ok:     false,
					Reason: "grm only accepts tags",
				},
				plumbing.ReferenceName("refs/tags/v1.0.0"): ReportItem{
					Ok:     true,
					Reason: "",
				},
				plumbing.ReferenceName("refs/tags/v1.0.1"): ReportItem{
					Ok:     true,
					Reason: "",
				},
				plumbing.ReferenceName("refs/tags/v1.1.0"): ReportItem{
					Ok:     true,
					Reason: "",
				},
			},
			wantW: "000eunpack ok\n002fng refs/heads/master grm only accepts tags\n0018ok refs/tags/v1.0.0\n0018ok refs/tags/v1.0.1\n0018ok refs/tags/v1.1.0\n0000",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				w := &bytes.Buffer{}
				tt.r.Write(w)
				if gotW := w.String(); gotW != tt.wantW {
					t.Errorf("Write() = %v, want %v", gotW, tt.wantW)
				}
			},
		)
	}
}
