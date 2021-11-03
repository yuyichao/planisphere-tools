package cmdr

import (
	"os"
)

type Slurper interface {
	Slurp(string) ([]byte, error)
}

type RealSlurper struct{}

func (c RealSlurper) Slurp(filepath string) ([]byte, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return b, err
}
