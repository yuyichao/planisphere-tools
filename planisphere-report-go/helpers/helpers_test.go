package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
)

func TestHashString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "test1",
			s:    "test1",
			want: "sha256:1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1d83567014",
		},
	}

	for _, tt := range tests {
		got := helpers.HashString(tt.s)
		require.Equal(t, tt.want, got)
	}
}
