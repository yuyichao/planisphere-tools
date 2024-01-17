/*
Package lookups defines the OSLookup interface and other accompanying bits
*/
package lookups

import (
	"context"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
	report "gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go"
)

// CheckedItems      []string
var checkedItemsMutex sync.Mutex

// OSLookuper defines the interface needed to lookup OS info
type OSLookuper interface {
	ApplyPlatformDetections(l *Lookup) error
	GetHostname(l *Lookup) (interface{}, error)
	GetInstalledSoftware(l *Lookup) (interface{}, error)
	GetOSFullName(l *Lookup) (interface{}, error)
	GetModel(l *Lookup) (interface{}, error)
	GetDeviceType(l *Lookup) (interface{}, error)
	GetOSFamily(l *Lookup) (interface{}, error)
	GetMemory(l *Lookup) (interface{}, error)
	GetDiskEncrypted(l *Lookup) (interface{}, error)
	GetManufacturer(l *Lookup) (interface{}, error)
	GetSerial(l *Lookup) (interface{}, error)
	GetExternalOSIdentifiers(l *Lookup) (interface{}, error)
}

// Lookup is the generic thing that does OS lookups
type Lookup struct {
	Overrides    map[string]interface{}
	Payload      planisphere.SelfReportPayload
	Commander    report.Commander
	CheckedItems []string
}

// LookupConfig is the configuration for a new Lookup item
type LookupConfig struct {
	Overrides map[string]interface{}
	OS        string // darwin, linux, windows, etc
	// CLI Interface for test mocking
	Commander *report.Commander
}

// WaitForChecked pauses the lookups until a specific item has been checked
func (l *Lookup) WaitForChecked(item string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		// if !util.ContainsString(l.CheckedItems, item) {
		if !slices.Contains(l.CheckedItems, item) {
			ci := l.CheckedItems
			sort.Strings(ci)
			slog.Debug("waiting on check", "item", item, "checked", ci)
			time.Sleep(1 * time.Second)
			if ctx.Err() != nil {
				slog.Warn("timed out waiting for check", "item", item)
				break
			}
		} else {
			break
		}
	}
}

// MarkChecked marks an item as checked
func (l *Lookup) MarkChecked(i string) {
	checkedItemsMutex.Lock()
	l.CheckedItems = append(l.CheckedItems, i)
	checkedItemsMutex.Unlock()
	slog.Debug("marked as checked", "item", i)
}
