package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Set up some standard errors for helers
var (
	ErrEmptyOutput        = fmt.Errorf("Empty output")
	ErrMalformedRPMOutput = fmt.Errorf("Malformed RPM Output")
)

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

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Wait for all items in wi to exist before continuing
func WaitForChecked(item string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		if !ContainsString(CheckedItems, item) {
			ci := CheckedItems
			sort.Strings(ci)
			log.Debugf("Waiting for %v to be checked...so far found: %v", item, ci)
			if ctx.Err() != nil {
				log.Warningf("Timed out waiting for %v to be checked", item)
				break
			}
		} else {
			break
		}
	}
}

func MarkChecked(i string) {
	log.Debugf("Marked %v as checked", i)
	checkedItemsMutex.Lock()
	CheckedItems = append(CheckedItems, i)
	checkedItemsMutex.Unlock()
}

func ParseRPMOutput(rpmOut []byte) ([][]string, error) {
	softwareTable := [][]string{}
	lines := []string{}
	trimmed := strings.Trim(string(rpmOut), "\n")
	// RPM can run successfully, but output nothing. Handle this by checking for a new line for now. See:
	// https://gitlab.oit.duke.edu/devil-ops/planisphere-tools/-/issues/3 for more info
	if strings.Contains(trimmed, "\n") {
		// Consider empty if just a new line return
		lines = strings.Split(trimmed, "\n")
	} else {
		// Consider empty if completely blank
		if trimmed == "" {
			return nil, ErrEmptyOutput
		} else if trimmed == "\n" {
			return nil, ErrEmptyOutput
		}
		lines = append(lines, trimmed)
	}
	for _, line := range lines {
		pieces := strings.Split(line, " ")
		if len(pieces) != 2 {
			return nil, ErrMalformedRPMOutput
		}
		name := pieces[0]
		version := pieces[1]

		softwareTable = append(softwareTable, []string{name, version})
	}

	return softwareTable, nil
}
