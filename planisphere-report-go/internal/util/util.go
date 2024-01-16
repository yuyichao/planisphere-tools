/*
Package util is just random utility commands
*/
package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// MakeNamePro inserts 'Pro' after the first part of the string
// so "Ubuntu 18.04" becomes "Ubuntu Pro 18.04"
func MakeNamePro(s string) string {
	pieces := strings.Split(s, " ")
	return fmt.Sprintf("%s Pro %s", pieces[0], strings.Join(pieces[1:], " "))
}

// ContainsString checks if a slice contains a given string
//
// Deprecated: Use slices.Contains instead nowadays
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Exists checks if a file exists
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
