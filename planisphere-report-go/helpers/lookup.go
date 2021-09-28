package helpers

import (
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
)

var CheckedItems []string
var checkedItemsMutex sync.Mutex

/* Lookuper will do the more advanced lookups. Using a custom struct for this so we
don't have to make duplicate system calls to look at the system_profiler
*/
type Lookuper struct {
	Overrides map[string]interface{}
	Payload   planisphere.SelfReportPayload
	// Things we know we have looked up here
}

// Use this to hold custom platform functional stuff
// This will help us loop through and do things in parallel
type platformSetter struct {
	name      string
	wrapper   func(*Lookuper, func(fl *Lookuper) (interface{}, error)) error
	platformF func(*Lookuper) (interface{}, error)
}

type genericSetter struct {
	name    string
	wrapper func(*Lookuper)
}

type setter struct {
	name string
	// Simple wrapper for just setting a value if it exists
	wrapper func(*Lookuper)
	// Enhanced wrapper that also takes a function to do the setting
	wrapperE func(*Lookuper, func(fl *Lookuper) (interface{}, error)) error
	// Function to pass in to the enhanced wrapper
	wrapperEF func(*Lookuper) (interface{}, error)
}

func NewLookuper(overrides map[string]interface{}) (*Lookuper, error) {
	var err error
	l := &Lookuper{
		Overrides: overrides,
	}

	// Super generic bits here
	l.Payload.LastActive = time.Now()

	err = ApplyPlatformDetections(l)
	if err != nil {
		return nil, err
	}

	genericSetters := []setter{
		{"Model", nil, setModelWrapper, setModel},
		{"InstanceKey", setInstanceKeyWrapper, nil, nil},
		{"DepartmentKey", setDepartmentKeyWrapper, nil, nil},
		{"SupportGroupID", setSupportGroupIDWrapper, nil, nil},
		{"SupportGroupName", setSupportGroupNameWrapper, nil, nil},
		{"UsageType", setUsageTypeWrapper, nil, nil},
		{"Status", setStatusWrapper, nil, nil},
		{"MacAddressses", setMacAddressesWrapper, nil, nil},
		{"ExtraData", setExtraDataWrapper, nil, nil},
		{"InstalledSoftware", nil, setInstalledSoftwareWrapper, setInstalledSoftware},
		{"Serial", nil, setPlatformSerialWrapper, setSerial},
		{"Manufacturer", nil, setManufacturerWrapper, setManufacturer},
		{"DiskEncrypted", nil, setDiskEncryptedWrapper, setDiskEncrypted},
		{"Memory", nil, setMemoryWrapper, setMemory},
		{"OSFamily", nil, setOSFamilyWrapper, setOSFamily},
		{"OSFullName", nil, setOSFullNameWrapper, setOSFullName},
		{"DeviceType", nil, setDeviceTypeWrapper, setDeviceType},
		{"Username", nil, setUsernameWrapper, setUsername},
		{"Hostname", nil, setHostnameWrapper, setHostname},
	}

	var wg1 sync.WaitGroup
	wg1.Add(len(genericSetters))
	for _, gs := range genericSetters {
		go func(gs setter) {
			defer wg1.Done()
			if gs.wrapperE == nil {
				// Simple setter operations
				gs.wrapper(l)
				log.Debug("Ran", gs.name)
			} else {
				// Enhanced setter operations
				err := gs.wrapperE(l, gs.wrapperEF)
				if err != nil {
					log.Warning(err)
				}
			}
		}(gs)
	}
	wg1.Wait()

	// Return
	return l, nil
}

/*
This is pretty OS agnostic, but if we need to later, we can break it out in to
OS specific functions
*/
func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}

func setPlatformSerialWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("serial")
	if item, ok := l.Overrides["serial"]; ok {
		l.Payload.Data.Serial = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.Serial = item.(string)
		}
	}
	return nil
}

func setManufacturerWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("manufacturer")
	if item, ok := l.Overrides["manufacturer"]; ok {
		l.Payload.Data.Manufacturer = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.Manufacturer = item.(string)
		}
	}

	return nil

}

func setModelWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("model")
	if item, ok := l.Overrides["model"]; ok {
		l.Payload.Data.Model = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.Model = item.(string)
		}
	}

	return nil

}

func setDiskEncryptedWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("disk_encrypted")
	if item, ok := l.Overrides["disk_encrypted"]; ok {
		l.Payload.Data.DiskEncrypted = item.(bool)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.DiskEncrypted = item.(bool)
		}
	}

	return nil

}

// Memory
// "All aloooooone in the moooooon liiiiiight"
//   - 😺
func setMemoryWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("memory_mb")
	if item, ok := l.Overrides["memory_mb"]; ok {
		l.Payload.Data.MemoryMB = uint64(item.(int))
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.MemoryMB = item.(uint64)
		}
	}

	return nil

}

// Operating System Stuff
// "I don't have friends, I got Family"
//   - Dominic Toretto 🚗💨
func setOSFamilyWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("os_family")
	if item, ok := l.Overrides["os_family"]; ok {
		l.Payload.Data.OsFamily = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.OsFamily = item.(string)
		}
	}

	return nil

}

func setDeviceTypeWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("device_type")
	if item, ok := l.Overrides["device_type"]; ok {
		l.Payload.Data.DeviceType = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.DeviceType = item.(string)
		}
	}

	return nil

}

func setOSFullNameWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("os_fullname")
	if item, ok := l.Overrides["os_fullname"]; ok {
		l.Payload.Data.OsFullname = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.OsFullname = item.(string)
		}
	}

	return nil

}

func setUsernameWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("username")
	if item, ok := l.Overrides["username"]; ok {
		l.Payload.Data.Username = item.(string)
	} else {
		item, err := f(l)
		if err != nil {
			return err
		} else {
			l.Payload.Data.Username = item.(string)
		}
	}

	return nil

}

func setHostnameWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("hostname")

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

func setInstanceKeyWrapper(l *Lookuper) {
	defer MarkChecked("instance_key")
	if instanceKey, ok := l.Overrides["instance_key"]; ok {
		l.Payload.Key = instanceKey.(string)
	}
}
func setDepartmentKeyWrapper(l *Lookuper) {
	defer MarkChecked("department_key")
	if departmentKey, ok := l.Overrides["department_key"]; ok {
		l.Payload.Data.DepartmentKey = departmentKey.(string)
	}
}

func setSupportGroupIDWrapper(l *Lookuper) {
	defer MarkChecked("support_group_id")
	if supportGroupID, ok := l.Overrides["support_group_id"]; ok {
		l.Payload.Data.SupportGroupId = uint64(supportGroupID.(int))
	}
}
func setSupportGroupNameWrapper(l *Lookuper) {
	defer MarkChecked("support_group_name")
	if supportGroupName, ok := l.Overrides["support_group_name"]; ok {
		l.Payload.Data.SupportGroupName = supportGroupName.(string)
	}
}
func setUsageTypeWrapper(l *Lookuper) {
	defer MarkChecked("usage_type")

	// Usage Type
	if usageType, ok := l.Overrides["usage_type"]; ok {
		l.Payload.Data.UsageType = usageType.(string)
	}
}

func setStatusWrapper(l *Lookuper) {
	defer MarkChecked("status")
	// Status: deployed, rma, etc
	if status, ok := l.Overrides["status"]; ok {
		l.Payload.Data.Status = status.(string)
	}
}

// Mac Addresses Field
// "Most Dope"
//   - Mac Miller ✌️
func setMacAddressesWrapper(l *Lookuper) {
	defer MarkChecked("mac_addresses")

	if macAddresses, ok := l.Overrides["mac_addresses"]; ok {
		for _, item := range macAddresses.([]interface{}) {
			l.Payload.Data.MacAddresses = append(l.Payload.Data.MacAddresses, item.(string))
		}
	} else {
		macs, err := getMacAddr()
		if err != nil {
			log.Warning("Could not detect mac_addresses: ", err)
		}
		l.Payload.Data.MacAddresses = macs
	}
}

func setExtraDataWrapper(l *Lookuper) {
	defer MarkChecked("extra_data")
	// Extra data
	l.Payload.ExtraData = map[string]string{}
	if extraData, ok := l.Overrides["extra_data"]; ok {
		extras, ok := extraData.(map[string]interface{})
		if ok {
			for k, v := range extras {
				l.Payload.ExtraData[k] = v.(string)
			}
		} else {
			log.Warning("No extra_data to parse, yet extra_data section exists")
		}
	}
}

func setInstalledSoftwareWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {
	defer MarkChecked("installed_software")

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
		} else {
			l.Payload.Data.InstalledSoftware = item.([][]string)
		}
	}

	return nil

}
