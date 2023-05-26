package util_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		s []string
		e string
		r bool
	}{
		{[]string{"a", "b", "c"}, "a", true},
		{[]string{"a", "b", "c"}, "d", false},
	}

	for _, test := range tests {
		got := util.ContainsString(test.s, test.e)
		require.Equal(t, test.r, got)
	}
}

func TestExists(t *testing.T) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	notafile := fmt.Sprintf("%v-not-a-file", file.Name())
	require.True(t, util.Exists(file.Name()))
	require.False(t, util.Exists(notafile))
}

func TestMakeNamePro(t *testing.T) {
	tests := map[string]struct {
		given string
		want  string
	}{
		"simple": {
			given: "Ubuntu 18.04",
			want:  "Ubuntu Pro 18.04",
		},
	}
	for desc, tt := range tests {
		require.Equal(t, tt.want, util.MakeNamePro(tt.given), desc)
	}
}
