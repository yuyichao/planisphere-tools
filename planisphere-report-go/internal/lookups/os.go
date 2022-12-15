package lookups

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

// CheckedItems      []string
var checkedItemsMutex sync.Mutex

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
	GetExternalOSIdentifiers(l *Lookuper) (interface{}, error)
}

type Lookuper struct {
	Overrides    map[string]interface{}
	Payload      planisphere.SelfReportPayload
	Commander    cmdr.Commander
	CheckedItems []string
}

type LookuperConfig struct {
	Overrides map[string]interface{}
	OS        string // darwin, linux, windows, etc
	// CLI Interface for test mocking
	Commander *cmdr.Commander
}

func (l *Lookuper) WaitForChecked(item string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		if !util.ContainsString(l.CheckedItems, item) {
			ci := l.CheckedItems
			sort.Strings(ci)
			log.Debug().Msgf("Waiting for %v to be checked...so far found: %v", item, ci)
			time.Sleep(1 * time.Second)
			if ctx.Err() != nil {
				log.Warn().Str("item", item).Msg("Timed out waiting for check")
				break
			}
		} else {
			break
		}
	}
}

func (l *Lookuper) MarkChecked(i string) {
	checkedItemsMutex.Lock()
	l.CheckedItems = append(l.CheckedItems, i)
	checkedItemsMutex.Unlock()
	log.Debug().Str("item", i).Msg("Marked as checked")
}
