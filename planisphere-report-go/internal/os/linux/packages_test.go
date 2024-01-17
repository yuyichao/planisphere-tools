package linux

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPackageOutput(t *testing.T) {
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
		// Test a single line of output
		{
			out: []byte(`gpg-pubkey 92d31755-5a81ef2e`),
			expect: [][]string{
				{"gpg-pubkey", "92d31755-5a81ef2e"},
			},
		},
		// Test a tabbed output
		{
			out: []byte("gpg-pubkey\t92d31755-5a81ef2e"),
			expect: [][]string{
				{"gpg-pubkey", "92d31755-5a81ef2e"},
			},
		},
	}

	for _, test := range tests {
		o, err := ParsePackageOutput(test.out)
		require.NoError(t, err)
		require.Equal(t, test.expect, o)
	}
}

func TestPackageEmptyOutput(t *testing.T) {
	tests := []struct {
		out []byte
	}{
		// No output
		{[]byte(``)},
		// Single newline
		{[]byte("\n")},
	}

	for _, test := range tests {
		_, err := ParsePackageOutput(test.out)
		require.EqualError(t, err, ErrEmptyOutput.Error())
	}
}
