/*
Package freebsd defines how to interact with the FreeBSD os
*/
package freebsd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

var errDmidecode = fmt.Errorf("error running dmidecode, ensure it is installed and you are running as root")

// OSLookup is the lookup object for FreeBSD
type OSLookup struct{}

// ApplyPlatformDetections sets the platform detection bits
func (o OSLookup) ApplyPlatformDetections(_ *lookups.Lookup) error {
	// Do initializing bits here
	return nil
}

// GetSerial returns the serial number
func (o OSLookup) GetSerial(l *lookups.Lookup) (interface{}, error) {
	cmdPath, err := l.Commander.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up serial number")
	}
	cmdOut, err := l.Commander.Output(cmdPath, "-s", "system-serial-number")
	if err != nil {
		return "", errors.New("issue running dmidecode to get the serial number")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

// GetManufacturer returns the manufacturer
func (o OSLookup) GetManufacturer(l *lookups.Lookup) (interface{}, error) {
	cmdPath, err := l.Commander.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up manufacturer")
	}
	cmdOut, err := l.Commander.Output(cmdPath, "-s", "chassis-manufacturer")
	if err != nil {
		return "", errors.New("issue running dmidecode to get the chassis-manufacturer")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

// GetModel returns the model
func (o OSLookup) GetModel(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Output("dmidecode", "-s", "chassis-version")
	if err != nil {
		return "", errDmidecode
	}
	trimmed := strings.Trim(string(out), "\n")
	return trimmed, nil
}

// GetDiskEncrypted returns if the full disk encryption is in use
func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookup) (interface{}, error) {
	return false, errors.New("diskEncrypted not yet implemented")
}

// GetMemory returns the amount of memory on the host
func (o OSLookup) GetMemory(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Output("/sbin/sysctl", "-n", "vm.kmem_size")
	if err != nil {
		return 0, err
	}
	trimmed := strings.Trim(string(out), "\n")
	memory, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, err
	}
	memoryMB := uint64(memory / 1024 / 1024)

	return memoryMB, nil
}

// GetOSFamily returns the OSFamily
func (o OSLookup) GetOSFamily(_ *lookups.Lookup) (interface{}, error) {
	return "FreeBSD", nil
}

// GetDeviceType returns the device type
func (o OSLookup) GetDeviceType(l *lookups.Lookup) (interface{}, error) {
	l.WaitForChecked("serial")
	if strings.Contains(l.Payload.Data.Serial, "VMware") {
		return "vm", nil
	}
	return "", nil
}

// GetOSFullName returns the full name of the os
func (o OSLookup) GetOSFullName(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Output("/bin/freebsd-version")
	if err != nil {
		return "", errors.New("could not run freebsd-version successfully")
	}
	trimmed := strings.Trim(string(out), "\n")
	return fmt.Sprintf("FreeBSD %v", trimmed), nil
}

// BSDSoftware represents the software installed
type BSDSoftware struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// GetInstalledSoftware returns the installed software
func GetInstalledSoftware(l *lookups.Lookup) ([][]string, error) {
	softwareTable := [][]string{}

	cmdPath, err := l.Commander.LookPath("pkg")
	if err == nil {
		cmdOut, err := l.Commander.Output(cmdPath, "info", "--raw", "-a", "--raw-format", "json-compact")
		if err != nil {
			slog.Warn("could not do a pkg info, even though the pkg command exists")
		}
		trimmed := strings.Trim(string(cmdOut), "\n")

		for _, line := range strings.Split(trimmed, "\n") {
			var item BSDSoftware
			pkgB := []byte(line)
			err = json.Unmarshal(pkgB, &item)
			if err != nil {
				return nil, err
			}

			name := item.Name
			version := item.Version

			softwareTable = append(softwareTable, []string{name, version})
		}
	} else {
		slog.Info("no pkg command installed")
	}

	return softwareTable, nil
}

// GetInstalledSoftware fulfills the OSLookuper interface
func (o OSLookup) GetInstalledSoftware(l *lookups.Lookup) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// GetHostname fulfills the OSLookuper interface
func (o OSLookup) GetHostname(_ *lookups.Lookup) (interface{}, error) {
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return hostname, nil
}

// GetExternalOSIdentifiers fulfills the OSLookuper interface
func (o OSLookup) GetExternalOSIdentifiers(_ *lookups.Lookup) (interface{}, error) {
	return nil, nil
}
