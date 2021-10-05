package helpers

import (
	"context"
	"errors"
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
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
