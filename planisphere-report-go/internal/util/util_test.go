package util_test

import (
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
