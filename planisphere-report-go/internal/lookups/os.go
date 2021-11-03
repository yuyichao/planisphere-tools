package lookups

import (
	"context"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

var (
	CheckedItems      []string
	checkedItemsMutex sync.Mutex
)

type OSLookup interface {
	ApplyPlatformDetections(l *Lookuper) error
	GetHostname(l *Lookuper) (interface{}, error)
	GetInstalledSoftware(l *Lookuper) (interface{}, error)
	GetOSFullName(l *Lookuper) (interface{}, error)
	GetModel(l *Lookuper) (interface{}, error)
	GetDeviceType(l *Lookuper) (interface{}, error)
	GetOSFamily(l *Lookuper) (interface{}, error)
	GetMemory(l *Lookuper) (interface{}, error)
	GetDiskEncrypted(l *Lookuper) (interface{}, error)
	GetManufacturer(l *Lookuper) (interface{}, error)
	GetSerial(l *Lookuper) (interface{}, error)
}

type Lookuper struct {
	Overrides map[string]interface{}
	Payload   planisphere.SelfReportPayload
	Commander cmdr.Commander
	Slurper   cmdr.Slurper
}

type LookuperConfig struct {
	Overrides map[string]interface{}
	OS        string // darwin, linux, windows, etc
	// CLI Interface for test mocking
	Commander *cmdr.Commander
	Slurper   *cmdr.Slurper
}

// Wait for all items in wi to exist before continuing
func WaitForChecked(item string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		if !util.ContainsString(CheckedItems, item) {
			ci := CheckedItems
			sort.Strings(ci)
			log.Debugf("Waiting for %v to be checked...so far found: %v", item, ci)
			time.Sleep(1 * time.Second)
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
