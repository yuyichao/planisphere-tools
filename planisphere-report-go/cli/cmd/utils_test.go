package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
		got := hashString(tt.s)
		require.Equal(t, tt.want, got)
	}
}

func TestExists(t *testing.T) {
	file, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	notafile := fmt.Sprintf("%v-not-a-file", file.Name())
	require.True(t, exists(file.Name()))
	require.False(t, exists(notafile))
}
