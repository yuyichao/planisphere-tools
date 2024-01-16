/*
Package linux defines how to interact with the Linux os stuff
*/
package linux

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/hardware"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
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

// GetInstalledSoftware returns the installed software
func GetInstalledSoftware(l *lookups.Lookup) ([][]string, error) {
	softwareTable := [][]string{}
	softwareQueries := [][]string{
		{"rpm", "-qa", "--qf", "%{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n"},
		{"guix-installed"},
		{"dpkg-query", "-W"},
		{"pacman", "-Q"},
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

// GetSerial satisfies the OSLookuper interface
func (o OSLookup) GetSerial(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Slurp("/sys/class/dmi/id/product_serial")
	if err != nil {
		return nil, err
	}
	serial := strings.TrimSpace(string(out))
	return serial, nil
}

// GetManufacturer satisfies the OSLookuper interface
func (o OSLookup) GetManufacturer(l *lookups.Lookup) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			return "Raspberry Pi", nil
		}
	}

	// bios_vendor isn't truly accurate on the linux hosts, moving to
	// sys_vendor which seems to be more inline with the other EPM tools
	out, err := l.Commander.Slurp("/sys/class/dmi/id/sys_vendor")
	if err != nil {
		return nil, err
	}
	vendor := strings.TrimSuffix(string(out), "\n")
	return vendor, nil
}

// GetModel satisfies the OSLookuper interface
func (o OSLookup) GetModel(l *lookups.Lookup) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			cpuDat, err := l.Commander.Slurp("/proc/cpuinfo")
			if err != nil {
				slog.Warn("could not get Raspberry Pi CPU info", "error", err)
				continue
			}
			trimmed := strings.Trim(string(cpuDat), "\n")
			for _, line := range strings.Split(trimmed, "\n") {
				pieces := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(pieces[0])
				value := strings.TrimSpace(pieces[1])
				if key == "Revision" {
					if _, ok := hardware.RaspberryPiModels[value]; ok {
						return hardware.RaspberryPiModels[value], nil
					}
				}
			}
		}
	}
	out, err := l.Commander.Slurp("/sys/class/dmi/id/product_name")
	if err != nil {
		return nil, err
	}
	productName := strings.TrimSuffix(string(out), "\n")
	return productName, nil
}

// GetDiskEncrypted satisfies the OSLookuper interface
func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookup) (interface{}, error) {
	return false, errors.New("DiskEncrypted Not yet implemented")
}

// GetMemory satisfies the OSLookuper interface
func (o OSLookup) GetMemory(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Slurp("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	reMemTotal := regexp.MustCompile(`^MemTotal:\s+(\d+)\s+.*$`)
	var memory uint64
	for s.Scan() {
		if m := reMemTotal.FindStringSubmatch(s.Text()); m != nil {
			memory, err = strconv.ParseUint(m[1], 10, 64)
			if err != nil {
				return nil, err
			}
		}
	}
	// Switch to MB here
	return memory / 1024, err
}

// GetOSFamily satisfies the OSLookuper interface
func (o OSLookup) GetOSFamily(_ *lookups.Lookup) (interface{}, error) {
	return "Linux", nil
}

// GetDeviceType satisfies the OSLookuper interface
func (o OSLookup) GetDeviceType(l *lookups.Lookup) (interface{}, error) {
	l.WaitForChecked("serial")
	if strings.Contains(l.Payload.Data.Serial, "VMware") {
		return "vm", nil
	}
	ct, err := l.Commander.Slurp("/sys/class/dmi/id/chassis_type")
	if err != nil {
		return nil, err
	}
	// Strip the newline please
	ctt := strings.TrimSpace(string(ct))
	// ctt as an integer
	cti, err := strconv.Atoi(ctt)
	if err != nil {
		return nil, err
	}
	// Is the chassis type integer in the deviceTypeTable?
	if _, ok := hardware.ChassisType[cti]; ok {
		return hardware.ChassisType[cti], nil
	}

	return "", nil
}

// GetOSFullName satisfies the OSLookuper interface
func (o OSLookup) GetOSFullName(l *lookups.Lookup) (interface{}, error) {
	osb, err := l.Commander.Slurp("/etc/os-release")
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(bytes.NewReader(osb))
	var fullName string

	rePrettyName := regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil {
			fullName = strings.Trim(m[1], `"`)
		}
	}
	if strings.HasPrefix(fullName, "Ubuntu") {
		if o.detectPro(l) || o.detectOITPro() {
			fullName = util.MakeNamePro(fullName)
		}
	}
	return fullName, nil
}

// detectPro attempts to determine if the ESM repos indicate that this server is
// getting full 'Pro' support
func (o OSLookup) detectPro(l *lookups.Lookup) bool {
	out, err := l.Commander.Output("/usr/bin/pro", "security-status", "--esm-infra", "--format", "json")
	if err != nil {
		// No pro yo
		return false
	}
	var esm esmStatus
	err = json.Unmarshal(out, &esm)
	if err != nil {
		return false
	}
	return esm.Summary.UA.Attached
}

// detectOITPro attempts to determin if the ESM repos are included in an OIT managed host
// OIT mirrors the pro repos locally instead of reaching out to the Ubuntu hosted packages.
// Because of this, the pro command incorrectly reports that it is not attached.
// We are instead checking to see if the local mirror repo exists, and assuming 'Pro' if
// it's there.
func (o OSLookup) detectOITPro() bool {
	matches, err := filepath.Glob("/etc/apt/sources.list.d/*-infra-updates.list")
	if err != nil {
		slog.Warn("error checking for OIT pro repos", "error", err)
	}
	if len(matches) > 0 {
		return true
	}
	return false
}

// esmStatus is a minimal holder for the esm status
type esmStatus struct {
	Summary struct {
		UA struct {
			Attached bool `json:"attached"`
		} `json:"ua"`
	} `json:"summary"`
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
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return hostname, nil
}

// GetExternalOSIdentifiers satisfies the OSLookuper interface
func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookup) (interface{}, error) {
	ids := map[string]string{}
	aidOut, err := l.Commander.Output("/opt/CrowdStrike/falconctl", "-g", "--aid")
	if err != nil {
		return nil, err
	}
	ids["crowdstrike_aid"] = strings.TrimSuffix(strings.TrimPrefix(string(aidOut), "aid=\""), "\".\n")
	return ids, nil
}
