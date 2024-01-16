package cmd

import (
	"crypto/sha256"
	"fmt"
)

func hashString(s string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(s)))
}
