package helpers

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/darwin"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/freebsd"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/linux"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

/* Lookuper will do the more advanced lookups. Using a custom struct for this so we
don't have to make duplicate system calls to look at the system_profiler
*/

type setter struct {
	name string
	// Simple wrapper for just setting a value if it exists
	wrapper func(*lookups.Lookuper)
	// Enhanced wrapper that also takes a function to do the setting
	wrapperE func(*lookups.Lookuper, func(fl *lookups.Lookuper) (interface{}, error)) error
	// Function to pass in to the enhanced wrapper
	wrapperEF func(*lookups.Lookuper) (interface{}, error)
}

func NewLookuper(c *lookups.LookuperConfig) (*lookups.Lookuper, error) {
	// var err error
	l := &lookups.Lookuper{
		Overrides: c.Overrides,
	}

	if c.Commander == nil {
		l.Commander = cmdr.RealCommander{}
	} else {
		l.Commander = *c.Commander
	}
	// Super generic bits here
	l.Payload.LastActive = time.Now()

	// Look up based on a given os, or detect runtime OS
	var fancyLookup lookups.OSLookup
	var detectOS string
	if c.OS == "" {
		detectOS = runtime.GOOS
	} else {
		detectOS = c.OS
	}
	slog.Debug("detected os", "os", detectOS)
	switch detectOS {
	case "darwin":
		fancyLookup = darwin.OSLookup{}
	case "linux":
		fancyLookup = linux.OSLookup{}
	case "freebsd":
		fancyLookup = freebsd.OSLookup{}
	default:
		return nil, fmt.Errorf("OS %s not supported", detectOS)
	}

	err := fancyLookup.ApplyPlatformDetections(l)
	if err != nil {
		return nil, err
	}

	genericSetters := []setter{
		{"Model", nil, setModelWrapper, fancyLookup.GetModel},
		{"ExternalOSIdentifiers", nil, setExternalOSIdentifersWrapper, fancyLookup.GetExternalOSIdentifiers},
		{"Memory", nil, setMemoryWrapper, fancyLookup.GetMemory},
		{"Serial", nil, setPlatformSerialWrapper, fancyLookup.GetSerial},
		{"Manufacturer", nil, setManufacturerWrapper, fancyLookup.GetManufacturer},
		{"DeviceType", nil, setDeviceTypeWrapper, fancyLookup.GetDeviceType},
		{"DiskEncrypted", nil, setDiskEncryptedWrapper, fancyLookup.GetDiskEncrypted},
		{"UsageType", setUsageTypeWrapper, nil, nil},
		{"Username", setUsernameWrapper, nil, nil},
		{"OSFamily", nil, setOSFamilyWrapper, fancyLookup.GetOSFamily},
		{"Status", setStatusWrapper, nil, nil},
		{"DepartmentKey", setDepartmentKeyWrapper, nil, nil},
		{"SupportGroupID", setSupportGroupIDWrapper, nil, nil},
		{"SupportGroupName", setSupportGroupNameWrapper, nil, nil},
		{"InstanceKey", setInstanceKeyWrapper, nil, nil},
		{"InstalledSoftware", nil, setInstalledSoftwareWrapper, fancyLookup.GetInstalledSoftware},
		{"ExtraData", setExtraDataWrapper, nil, nil},
		{"Hostname", nil, setHostnameWrapper, fancyLookup.GetHostname},
		{"OSFullName", nil, setOSFullNameWrapper, fancyLookup.GetOSFullName},
		{"MacAddressses", setMacAddressesWrapper, nil, nil},
	}

	var wg1 sync.WaitGroup
	wg1.Add(len(genericSetters))
	for _, gs := range genericSetters {
		go func(gs setter) {
			defer wg1.Done()
			if gs.wrapperE == nil {
				// Simple setter operations
				gs.wrapper(l)
				slog.Debug("ran setter", "setter", gs.name)
			} else {
				// Enhanced setter operations
				err := gs.wrapperE(l, gs.wrapperEF)
				if err != nil {
					slog.Warn("error running setter", "error", err)
				}
			}
		}(gs)
	}
	wg1.Wait()

	return l, nil
}

func setPlatformSerialWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("serial")
	if item, ok := l.Overrides["serial"]; ok {
		l.Payload.Data.Serial = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.Serial = item.(string)
	}
	return nil
}

func setManufacturerWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("manufacturer")
	if item, ok := l.Overrides["manufacturer"]; ok {
		l.Payload.Data.Manufacturer = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.Manufacturer = item.(string)
	}

	return nil
}

func setModelWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("model")
	if item, ok := l.Overrides["model"]; ok {
		l.Payload.Data.Model = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.Model = item.(string)
	}

	return nil
}

func setDiskEncryptedWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("disk_encrypted")
	if item, ok := l.Overrides["disk_encrypted"]; ok {
		l.Payload.Data.DiskEncrypted = item.(bool)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.DiskEncrypted = item.(bool)
	}

	return nil
}

// Memory
// "All aloooooone in the moooooon liiiiiight"
//   - 😺
func setMemoryWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("memory_mb")
	if item, ok := l.Overrides["memory_mb"]; ok {
		l.Payload.Data.MemoryMB = uint64(item.(int))
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.MemoryMB = item.(uint64)
	}

	return nil
}

// Operating System Stuff
// "I don't have friends, I got Family"
//   - Dominic Toretto 🚗💨
func setOSFamilyWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("os_family")
	if item, ok := l.Overrides["os_family"]; ok {
		l.Payload.Data.OsFamily = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.OsFamily = item.(string)
	}

	return nil
}

func setDeviceTypeWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("device_type")
	var t string
	if item, ok := l.Overrides["device_type"]; ok {
		t = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		t = item.(string)
	}
	validTypes := []string{"desktop", "laptop", "server", "server_physical", "vm"}
	if util.ContainsString(validTypes, t) {
		l.Payload.Data.DeviceType = t
		return nil
	}
	slog.Warn("invalid device type", "device_type", t, "valid_types", validTypes)
	return errors.New("InvalidDeviceType")
}

func setOSFullNameWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("os_fullname")
	if item, ok := l.Overrides["os_fullname"]; ok {
		l.Payload.Data.OsFullname = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.OsFullname = item.(string)
	}

	return nil
}

func setHostnameWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("hostname")

	// Hostname Field
	if hostname, ok := l.Overrides["hostname"]; ok {
		l.Payload.Data.Hostname = hostname.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.Hostname = item.(string)
	}
	return nil
}

func setInstanceKeyWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("instance_key")
	if instanceKey, ok := l.Overrides["instance_key"]; ok {
		l.Payload.Key = instanceKey.(string)
	}
}

func setDepartmentKeyWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("department_key")
	if departmentKey, ok := l.Overrides["department_key"]; ok {
		l.Payload.Data.DepartmentKey = departmentKey.(string)
	}
}

func setSupportGroupIDWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("support_group_id")
	if supportGroupID, ok := l.Overrides["support_group_id"]; ok {
		l.Payload.Data.SupportGroupId = uint64(supportGroupID.(int))
	}
}

func setSupportGroupNameWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("support_group_name")
	if supportGroupName, ok := l.Overrides["support_group_name"]; ok {
		l.Payload.Data.SupportGroupName = supportGroupName.(string)
	}
}

func setUsageTypeWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("usage_type")

	// Usage Type
	if usageType, ok := l.Overrides["usage_type"]; ok {
		l.Payload.Data.UsageType = usageType.(string)
	}
}

func setUsernameWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("username")

	if usageType, ok := l.Overrides["username"]; ok {
		l.Payload.Data.Username = usageType.(string)
	}
}

func setStatusWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("status")
	// Status: deployed, rma, etc
	if status, ok := l.Overrides["status"]; ok {
		l.Payload.Data.Status = status.(string)
	}
}

// Mac Addresses Field
// "Most Dope"
//   - Mac Miller ✌️
func setMacAddressesWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("mac_addresses")

	if macAddresses, ok := l.Overrides["mac_addresses"]; ok {
		for _, item := range macAddresses.([]interface{}) {
			l.Payload.Data.MacAddresses = append(l.Payload.Data.MacAddresses, item.(string))
		}
	} else {
		// macs, err := getMacAddr()
		macs, err := l.Commander.GetMacAddrs()
		if err != nil {
			slog.Warn("could not detect mac_addresses", "error", err)
		}
		l.Payload.Data.MacAddresses = macs
	}
}

func setExtraDataWrapper(l *lookups.Lookuper) {
	defer l.MarkChecked("extra_data")
	// Extra data
	l.Payload.ExtraData = map[string]string{}
	if extraData, ok := l.Overrides["extra_data"]; ok {
		extras, ok := extraData.(map[string]interface{})
		if ok {
			for k, v := range extras {
				l.Payload.ExtraData[k] = v.(string)
			}
		} else {
			slog.Warn("no extra_data to parse, yet extra_data section exists")
		}
	}
}

func setInstalledSoftwareWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("installed_software")

	// Not sure why someone would wanna override this, but just in case...
	if installedSoftware, ok := l.Overrides["installed_software"]; ok {
		for _, item := range installedSoftware.([]interface{}) {
			pieces := []string{}
			for _, piece := range item.([]interface{}) {
				pieces = append(pieces, piece.(string))
			}
			itemPair := []string{pieces[0], pieces[1]}
			l.Payload.Data.InstalledSoftware = append(l.Payload.Data.InstalledSoftware, itemPair)
		}
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		l.Payload.Data.InstalledSoftware = item.([][]string)
	}

	return nil
}

func setExternalOSIdentifersWrapper(l *lookups.Lookuper, f func(fl *lookups.Lookuper) (interface{}, error)) error {
	defer l.MarkChecked("external_os_identifiers")
	// Status: deployed, rma, etc
	if item, ok := l.Overrides["external_os_identifiers"]; ok {
		l.Payload.Data.ExternalOSIdentifiers = item.(map[string]string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		}
		if item != nil {
			l.Payload.Data.ExternalOSIdentifiers = item.(map[string]string)
		}
	}
	return nil
}
