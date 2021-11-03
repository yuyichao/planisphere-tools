package linux_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/linux"
)

func TestRPMOutput(t *testing.T) {
	tests := []struct {
		out    []byte
		expect [][]string
	}{
		// Test a multi-line output
		{
			out: []byte(`gpg-pubkey 92d31755-5a81ef2e
device-mapper-libs 8:1.02.177-10.el8`),
			expect: [][]string{
				{"gpg-pubkey", "92d31755-5a81ef2e"},
				{"device-mapper-libs", "8:1.02.177-10.el8"},
			},
		},
		// Test a single line of RPM output
		{
			out: []byte(`gpg-pubkey 92d31755-5a81ef2e`),
			expect: [][]string{
				{"gpg-pubkey", "92d31755-5a81ef2e"},
			},
		},
	}

	for _, test := range tests {
		o, err := linux.ParseRPMOutput(test.out)
		require.NoError(t, err)
		require.Equal(t, test.expect, o)
	}
}

func TestRPMEmptyOutput(t *testing.T) {
	tests := []struct {
		out []byte
	}{
		// No output
		{[]byte(``)},
		// Single newline
		{[]byte("\n")},
	}

	for _, test := range tests {
		_, err := linux.ParseRPMOutput(test.out)
		require.EqualError(t, err, linux.ErrEmptyOutput.Error())
	}
}
