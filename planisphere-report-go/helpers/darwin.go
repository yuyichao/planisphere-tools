// go:build darwin
package helpers

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
Only platform specific stuff should be in here. If it's more generic than a
given platform, please include it in the NewLookup function
*/
func ApplyPlatformDetections(l *Lookuper) error {

	ioregExpert, err := GetIORegTree("IOPlatformExpertDevice")
	if err != nil {
		return err
	}

	// Disk encryption state
	if encrypted, ok := l.Overrides["disk_encrypted"]; ok {
		l.Payload.Data.DiskEncrypted = encrypted.(bool)
	} else {
		encrypted, err := GetDiskEncryptionStatus()
		if err != nil {
			log.Warning("Could not detect disk encryption state: ", err)
		}
		l.Payload.Data.DiskEncrypted = encrypted
	}

	// Memory
	// "All aloooooone in the moooooon liiiiiight"
	//   - 😺
	if memory, ok := l.Overrides["memory_mb"]; ok {
		m := int64(memory.(int))

		l.Payload.Data.MemoryMB = uint64(m)
	} else {
		memory, err := GetMemory()
		if err != nil {
			log.Warning("Could not detect memory")
		}
		l.Payload.Data.MemoryMB = uint64(memory)
	}

	// Detect Serial Number
	if serial, ok := l.Overrides["serial"]; ok {
		l.Payload.Data.Serial = serial.(string)
	} else {
		l.Payload.Data.Serial = ioregExpert["IOPlatformSerialNumber"]
	}

	// Operating System Stuff
	// "I don't have friends, I got Family"
	//   - Dominic Toretto 🚗💨
	if osFamily, ok := l.Overrides["os_family"]; ok {
		l.Payload.Data.OsFamily = osFamily.(string)
	} else {
		l.Payload.Data.OsFamily = "macOS"
	}

	// Manufacturer
	if manufacturer, ok := l.Overrides["manufacturer"]; ok {
		l.Payload.Data.Manufacturer = manufacturer.(string)
	} else {
		l.Payload.Data.Manufacturer = ioregExpert["manufacturer"]
	}

	// Model
	if model, ok := l.Overrides["model"]; ok {
		l.Payload.Data.Model = model.(string)
	} else {
		l.Payload.Data.Model = ioregExpert["product-name"]
	}

	// What kind of device is this?
	// TODO: Flesh this out more
	if deviceType, ok := l.Overrides["model"]; ok {
		l.Payload.Data.DeviceType = deviceType.(string)
	} else {
		if strings.Contains(l.Payload.Data.Model, "MacBook") {
			l.Payload.Data.DeviceType = "laptop"
		}
	}

	// OS Data here
	softData, err := GetPSoftwareData()
	if err != nil {
		return err
	}

	if osVersion, ok := l.Overrides["os_version"]; ok {
		l.Payload.Data.OsFullname = osVersion.(string)
	} else {
		l.Payload.Data.OsFullname = softData.SPSoftwareDataType[0].OsVersion
	}

	// Username, cool to override
	if username, ok := l.Overrides["username"]; ok {
		l.Payload.Data.Username = username.(string)
	} else {
		l.Payload.Data.Username = softData.SPSoftwareDataType[0].UserName
	}

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
		apps, err := GetInstalledSoftware()
		if err != nil {
			return err
		}
		l.Payload.Data.InstalledSoftware = apps
	}

	// When last active?
	l.Payload.LastActive = time.Now()

	// Mmmm, return data
	return nil
}

type SPSoftwareData struct {
	SPSoftwareDataType []struct {
		Name            string `json:"_name,omitempty"`
		BootMode        string `json:"boot_mode,omitempty"`
		BootVolume      string `json:"boot_volume,omitempty"`
		KernelVersion   string `json:"kernel_version,omitempty"`
		LocalHostName   string `json:"local_host_name,omitempty"`
		OsVersion       string `json:"os_version,omitempty"`
		SecureVm        string `json:"secure_vm,omitempty"`
		SystemIntegrity string `json:"system_integrity,omitempty"`
		Uptime          string `json:"uptime,omitempty"`
		UserName        string `json:"user_name,omitempty"`
	}
}

type SPApplicationData struct {
	SPApplicationsDataType []struct {
		Name         string   `json:"_name,omitempty"`
		ArchKind     string   `json:"arch_kind,omitempty"`
		LastModified string   `json:"lastModified,omitempty"`
		ObtainedFrom string   `json:"obtained_from,omitempty"`
		Path         string   `json:"path,omitempty"`
		SignedBy     []string `json:"signed_by,omitempty"`
		Version      string   `json:"version,omitempty"`
	}
}

// This is like...OS data dude...
func GetPSoftwareData() (SPSoftwareData, error) {
	var s SPSoftwareData
	out, err := exec.Command("/usr/sbin/system_profiler", "SPSoftwareDataType", "-json").Output()
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(out, &s)
	if err != nil {
		return s, err
	}
	return s, nil

}

// This is like...Application level data...my cool person
func GetPSApplicationData() (SPApplicationData, error) {
	var s SPApplicationData
	out, err := exec.Command("/usr/sbin/system_profiler", "SPApplicationsDataType", "-json").Output()
	if err != nil {
		log.Warning("Error converting apps: ", err)
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(out, &s)
	if err != nil {
		return s, err
	}
	return s, nil

}

func GetMemory() (int64, error) {

	// Memory here
	memory, err := GetSysctl("hw.memsize")
	if err != nil {
		return 0, err
	}
	memoryMB := memory / 1024 / 1024

	return memoryMB, nil

}

func GetInstalledSoftware() ([][]string, error) {
	softwareTable := [][]string{}

	/*
		Homebrew packages reported here
	*/
	brewOut, err := exec.Command("/usr/local/bin/brew", "list", "--versions").Output()
	if err != nil {
		log.Warning("Homebrew package lookup failed")
	} else {
		trimmed := strings.Trim(string(brewOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, " ")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	}

	/*
		This is applications that MacOS knows about. The data is a little
		inconsistent as many packages don't list a version. Right now we are
		reporting the 'name', which may also be misleading. A more unique
		identifer might be 'path' for this...we should think about what the
		right way to report back is
	*/
	// Application Data
	data, err := GetPSApplicationData()
	if err != nil {
		return nil, err
	}
	if err != nil {
		log.Warning("Could not get app date: ", err)
		return nil, err
	}
	for _, item := range data.SPApplicationsDataType {
		softwareTable = append(softwareTable, []string{item.Name, item.Version})
	}

	return softwareTable, nil
}

func GetSysctl(target string) (int64, error) {
	out, err := exec.Command("/usr/sbin/sysctl", "-n", target).Output()
	if err != nil {
		return 0, err
	}
	outClean := strings.TrimSuffix(string(out), "\n")

	v, err := strconv.ParseInt(outClean, 10, 64)
	if err != nil {
		log.Warning("Error doing sysctl: ", err)
		return 0, err
	}
	return v, nil
}
func GetIORegValue(tree, item string) (string, error) {
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", tree).Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		stripLine := strings.TrimSpace(line)
		if !strings.HasPrefix(stripLine, "\"") {
			continue
		}
		pieces := strings.Split(stripLine, " = ")
		// Strip off head and tail "s
		key := pieces[0]
		key = strings.ReplaceAll(key, "\"", "")

		// Strip < > from value
		value := pieces[1]
		value = strings.TrimLeft(value, "<")
		value = strings.TrimRight(value, ">")
		value = strings.ReplaceAll(value, "\"", "")
		if key == item {
			return value, nil
		}
	}
	return "", nil

}

func GetIORegTree(tree string) (map[string]string, error) {
	r := make(map[string]string)
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", tree).Output()
	if err != nil {
		return r, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		stripLine := strings.TrimSpace(line)
		if !strings.HasPrefix(stripLine, "\"") {
			continue
		}
		pieces := strings.Split(stripLine, " = ")
		// Strip off head and tail "s
		key := pieces[0]
		key = strings.ReplaceAll(key, "\"", "")

		// Strip < > from value
		value := pieces[1]
		value = strings.TrimLeft(value, "<")
		value = strings.TrimRight(value, ">")
		value = strings.ReplaceAll(value, "\"", "")
		r[key] = value
	}
	return r, nil

}

func GetDiskEncryptionStatus() (bool, error) {
	out, err := exec.Command("/usr/bin/fdesetup", "isactive").Output()
	if err != nil {
		return false, err
	}
	outS := string(out)
	outS = strings.TrimRight(outS, "\n")
	if outS == "true" {
		return true, nil
	} else {
		return false, nil
	}
}
