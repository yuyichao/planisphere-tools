package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
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
		o, err := helpers.ParseRPMOutput(test.out)
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
		_, err := helpers.ParseRPMOutput(test.out)
		require.EqualError(t, err, helpers.ErrEmptyOutput.Error())
	}
}

func TestRPMMalformedOutput(t *testing.T) {
	tests := []struct {
		out []byte
	}{
		// Test lines without versions
		{out: []byte(`foo
bar`)},
		// Test lines with mixed versions and non-versions
		{out: []byte(`foo
bar v1.2.3`)},
	}

	for _, test := range tests {
		_, err := helpers.ParseRPMOutput(test.out)
		require.EqualError(t, err, helpers.ErrMalformedRPMOutput.Error())
	}
}
