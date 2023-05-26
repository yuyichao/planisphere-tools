package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// makeNamePro inserts 'Pro' after the first part of the string
// so "Ubuntu 18.04" becomes "Ubuntu Pro 18.04"
func MakeNamePro(s string) string {
	pieces := strings.Split(s, " ")
	primary := pieces[0]
	remainder := pieces[1:]
	return fmt.Sprintf("%s Pro %s", primary, strings.Join(remainder, " "))
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
