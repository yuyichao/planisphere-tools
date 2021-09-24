//go:build darwin
// +build darwin

package helpers

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
Only platform specific stuff should be in here. If it's more generic than a
given platform, please include it in the NewLookup function
Is there anything we actually want to do here globally?
*/
// ApplyPlatformDetections Do some stuff here
var ioregExpert map[string]string

func ApplyPlatformDetections(l *Lookuper) error {
	var err error
	ioregExpert, err = GetIORegTree("IOPlatformExpertDevice")
	if err != nil {
		return err
	}

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
	// Note fdsetup fails if the encryption is inactive
	_, err := exec.Command("/usr/bin/fdesetup", "isactive").Output()
	if err == nil {
		return true, nil
	} else {
		return false, nil
	}
}

func setSerial(l *Lookuper) (interface{}, error) {
	return ioregExpert["IOPlatformSerialNumber"], nil
}

func setManufacturer(l *Lookuper) (interface{}, error) {
	return ioregExpert["manufacturer"], nil

}

func setModel(l *Lookuper) (interface{}, error) {
	return ioregExpert["product-name"], nil

}

func setDiskEncrypted(l *Lookuper) (interface{}, error) {

	encrypted, err := GetDiskEncryptionStatus()
	if err != nil {
		log.Warning("Could not detect disk encryption state: ", err)
	}
	return encrypted, nil

}

func setMemory(l *Lookuper) (interface{}, error) {

	memory, err := GetMemory()
	if err != nil {
		log.Warning("Could not detect memory")
	}
	return uint64(memory), nil

}

func setOSFamily(l *Lookuper) (interface{}, error) {

	return "macOS", nil

}

func setDeviceType(l *Lookuper) (interface{}, error) {
	if strings.Contains(l.Payload.Data.Model, "MacBook") {
		return "laptop", nil
	} else {
		return "", errors.New("Unknown type of mac")
	}

}

func setOSFullName(l *Lookuper) (interface{}, error) {
	softData, err := GetPSoftwareData()
	if err != nil {
		return "", err
	}

	return softData.SPSoftwareDataType[0].OsVersion, nil

}

func setUsername(l *Lookuper) (interface{}, error) {
	/*
		softData, err := GetPSoftwareData()
		if err != nil {
			return "", err
		}
		return softData.SPSoftwareDataType[0].UserName, nil
	*/
	out, err := exec.Command("/usr/bin/last").Output()
	if err != nil {
		log.Warning("Error running 'last' to determine the real user")
	}
	for _, line := range strings.Split(string(out), "\n") {
		pieces := strings.Split(line, " ")
		for _, piece := range pieces {
			if piece != "root" {
				return piece, nil
			}
		}
	}
	return "", errors.New("Could not find a non-root user")

}

func setInstalledSoftware(l *Lookuper) (interface{}, error) {
	apps, err := GetInstalledSoftware()
	if err != nil {
		return nil, err
	}
	return apps, nil

}
