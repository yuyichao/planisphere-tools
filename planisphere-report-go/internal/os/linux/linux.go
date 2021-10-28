/*
Package linux defines how to interact with the Linux os stuff
*/
package linux

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

var (
	// ErrEmptyOutput is when there are no actual lines outputted
	ErrEmptyOutput = fmt.Errorf("empty output")
	// ErrMalformedPackageOutput is when the package output is bad
	ErrMalformedPackageOutput = fmt.Errorf("malformed Package Output")
)

// OSLookup handles the Linux OS lookups
type OSLookup struct{}

// ApplyPlatformDetections satisfies the OSLookuper interface
func (o OSLookup) ApplyPlatformDetections(_ *lookups.Lookup) error {
	return nil
}

// GetEnabledRepos returns a list of enabled repositories on a given system
func GetEnabledRepos(l *lookups.Lookup) ([]string, error) {
	return nil, errors.New("no apt-cache or yum found to look up enabled repositories")
}

// GetInstalledSoftware returns the installed software
func GetInstalledSoftware(l *lookups.Lookup) ([][]string, error) {
	softwareTable := [][]string{}
	softwareQueries := [][]string{
		{"pacman", "-Q", "linux", "glibc"},
	}

	for _, softwareQuery := range softwareQueries {
		cmd := softwareQuery[0]
		args := softwareQuery[1:]
		binPath, err := l.Commander.LookPath(cmd)
		if err == nil {
			slog.Debug("running software query", "query", softwareQuery)
			binOut, err := l.Commander.Output(binPath, args...)
			if err != nil {
				slog.Warn("could not do a listing even though the rpm command exists", "command", cmd)
			}
			binSoftware, err := ParsePackageOutput(binOut)
			if err != nil {
				slog.Warn("could not parse the command output", "command", cmd)
			} else {
				softwareTable = append(softwareTable, binSoftware...)
			}
		} else {
			slog.Debug("command installed", "cmd", cmd)
		}
	}

	return softwareTable, nil
}

// GetExtendedOSSupport returns the vendor providing Extended OS Support for an operating system
func (o OSLookup) GetExtendedOSSupport(l *lookups.Lookup) (interface{}, error) {
	return "", nil
}

// GetSerial satisfies the OSLookuper interface
func (o OSLookup) GetSerial(l *lookups.Lookup) (interface{}, error) {
	return "Serial", nil
}

// GetManufacturer satisfies the OSLookuper interface
func (o OSLookup) GetManufacturer(l *lookups.Lookup) (interface{}, error) {
	return "Vendor", nil
}

// GetModel satisfies the OSLookuper interface
func (o OSLookup) GetModel(l *lookups.Lookup) (interface{}, error) {
	return "Model", nil
}

// GetDiskEncrypted satisfies the OSLookuper interface
func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookup) (interface{}, error) {
	return false, errors.New("DiskEncrypted Not yet implemented")
}

// GetMemory satisfies the OSLookuper interface
func (o OSLookup) GetMemory(l *lookups.Lookup) (interface{}, error) {
	return uint64(1024), nil
}

// GetOSFamily satisfies the OSLookuper interface
func (o OSLookup) GetOSFamily(_ *lookups.Lookup) (interface{}, error) {
	return "Linux", nil
}

// GetDeviceType satisfies the OSLookuper interface
func (o OSLookup) GetDeviceType(l *lookups.Lookup) (interface{}, error) {
	return "", nil
}

// GetOSFullName satisfies the OSLookuper interface
func (o OSLookup) GetOSFullName(l *lookups.Lookup) (interface{}, error) {
	return "ArchLinux", nil
}

func aptRepos(l *lookups.Lookup) ([]string, error) {
	return nil, errors.New("could not run apt-cache")
}

func yumRepos(l *lookups.Lookup) ([]string, error) {
	return nil, errors.New("could not run yum")
}

// GetInstalledSoftware satisfies the OSLookuper interface
func (o OSLookup) GetInstalledSoftware(l *lookups.Lookup) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// GetHostname satisfies the OSLookuper interface
func (o OSLookup) GetHostname(_ *lookups.Lookup) (interface{}, error) {
	return "hostname", nil
}

// GetExternalOSIdentifiers satisfies the OSLookuper interface
func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookup) (interface{}, error) {
	return nil, errors.New("could not run crowdstrike")
}

// MakeNamePro inserts 'Pro' after the first part of the string
// so "Ubuntu 18.04" becomes "Ubuntu Pro 18.04"
func MakeNamePro(s string) string {
	pieces := strings.Split(s, " ")
	return fmt.Sprintf("%s Pro %s", pieces[0], strings.Join(pieces[1:], " "))
}
