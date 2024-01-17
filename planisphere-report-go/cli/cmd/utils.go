package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
)

func hashString(s string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(s)))
}

// exists checks if a file exists
func exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
