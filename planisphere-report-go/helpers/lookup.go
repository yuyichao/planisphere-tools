package helpers

import (
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
)

/*
Lookuper will do the more advanced lookups. Using a custom struct for this so we
don't have to make duplicate system calls to look at the system_profiler
*/
type Lookuper struct {
	Overrides map[string]interface{}
	Payload   planisphere.SelfReportPayload
}

func NewLookuper(overrides map[string]interface{}) (*Lookuper, error) {
	l := &Lookuper{
		Overrides: overrides,
	}
	// Here are some things that can only be set in the overrides section

	// Key - Not the API Key, but the instance key. This is the 'key' part of the post
	if instanceKey, ok := l.Overrides["instance_key"]; ok {
		l.Payload.Key = instanceKey.(string)
	}

	// Department Key
	if departmentKey, ok := l.Overrides["department_key"]; ok {
		l.Payload.Data.DepartmentKey = departmentKey.(string)
	}

	// Support Group ID
	if supportGroupID, ok := l.Overrides["support_group_id"]; ok {
		l.Payload.Data.SupportGroupId = uint64(supportGroupID.(int))
	}

	// Support Group Name
	if supportGroupName, ok := l.Overrides["support_group_name"]; ok {
		l.Payload.Data.SupportGroupName = supportGroupName.(string)
	}

	// Usage Type
	if usageType, ok := l.Overrides["usage_type"]; ok {
		l.Payload.Data.UsageType = usageType.(string)
	}

	// Status: deployed, rma, etc
	if status, ok := l.Overrides["status"]; ok {
		l.Payload.Data.Status = status.(string)
	}

	// Hostname Field
	if hostname, ok := l.Overrides["hostname"]; ok {
		l.Payload.Data.Hostname = hostname.(string)
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			log.Warning("Could not detect hostname: ", err)
		}
		l.Payload.Data.Hostname = hostname
	}

	// Mac Addresses Field
	// "Most Dope"
	//   - Mac Miller ✌️
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

	// Extra data
	l.Payload.ExtraData = map[string]string{}
	if extraData, ok := l.Overrides["extra_data"]; ok {
		for k, v := range extraData.(map[string]interface{}) {
			log.Println(k, v)
			l.Payload.ExtraData[k] = v.(string)
		}
	}

	// Now
	var err error
	l.Payload.LastActive = time.Now()

	err = ApplyPlatformDetections(l)
	if err != nil {
		return nil, err
	}

	l.Payload.LastActive = time.Now()

	/*
		platformItems := []func(*Lookuper, func(*Lookuper, interface{}) error){}
		platformItems = append(platformItems, setPlatformSerialWrapper(l, setSerial))
		log.Println(platformItems)
	*/
	//platformItems = append(platformItems, setPlatformSerialWrapper(l, setSerial))

	// Do the platformy stuff
	err = setPlatformSerialWrapper(l, setSerial)
	if err != nil {
		log.Warning(err)
	}
	err = setManufacturerWrapper(l, setManufacturer)
	if err != nil {
		log.Warning(err)
	}
	err = setModelWrapper(l, setModel)
	if err != nil {
		log.Warning(err)
	}
	err = setDiskEncryptedWrapper(l, setDiskEncrypted)
	if err != nil {
		log.Warning(err)
	}
	err = setMemoryWrapper(l, setMemory)
	if err != nil {
		log.Warning(err)
	}
	err = setOSFamilyWrapper(l, setOSFamily)
	if err != nil {
		log.Warning(err)
	}
	err = setOSFullNameWrapper(l, setOSFullName)
	if err != nil {
		log.Warning(err)
	}
	err = setDeviceTypeWrapper(l, setDeviceType)
	if err != nil {
		log.Warning(err)
	}
	err = setUsernameWrapper(l, setUsername)
	if err != nil {
		log.Warning(err)
	}
	err = setInstalledSoftwareWrapper(l, setInstalledSoftware)
	if err != nil {
		log.Warning(err)
	}

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

func setInstalledSoftwareWrapper(l *Lookuper, f func(fl *Lookuper) (interface{}, error)) error {

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
