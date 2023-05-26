package linux

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/hardware"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

var (
	ErrEmptyOutput            = fmt.Errorf("Empty output")
	ErrMalformedPackageOutput = fmt.Errorf("Malformed Package Output")
)

type OSLookup struct{}

func (o OSLookup) ApplyPlatformDetections(_ *lookups.Lookuper) error {
	return nil
}

func GetInstalledSoftware(l *lookups.Lookuper) ([][]string, error) {
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
			log.Debug().Strs("query", softwareQuery).Msg("Running software query")
			binOut, err := l.Commander.Output(binPath, args...)
			if err != nil {
				log.Warn().Str("command", cmd).Msg("Could not do alisting even though the rpm command exists")
			}
			binSoftware, err := ParsePackageOutput(binOut)
			if err != nil {
				log.Warn().Str("command", cmd).Msg("Could not parse the command output")
			} else {
				softwareTable = append(softwareTable, binSoftware...)
			}
		} else {
			log.Debug().Str("cmd", cmd).Msg("command installed")
		}
	}

	return softwareTable, nil
}

func (o OSLookup) GetSerial(l *lookups.Lookuper) (interface{}, error) {
	out, err := l.Commander.Slurp("/sys/class/dmi/id/product_serial")
	if err != nil {
		return nil, err
	}
	serial := strings.TrimSpace(string(out))
	return serial, nil
}

func (o OSLookup) GetManufacturer(l *lookups.Lookuper) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			return "Raspberry Pi", nil
		}
	}
	out, err := l.Commander.Slurp("/sys/class/dmi/id/bios_vendor")
	if err != nil {
		return nil, err
	}
	vendor := strings.TrimSuffix(string(out), "\n")
	return vendor, nil
}

func (o OSLookup) GetModel(l *lookups.Lookuper) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			cpuDat, err := l.Commander.Slurp("/proc/cpuinfo")
			if err != nil {
				log.Warn().Err(err).Msg("Could not get Raspberry Pi CPU info")
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

func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookuper) (interface{}, error) {
	// TODO: Implement this
	return false, errors.New("DiskEncrypted Not yet implemented")
}

func (o OSLookup) GetMemory(l *lookups.Lookuper) (interface{}, error) {
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

func (o OSLookup) GetOSFamily(_ *lookups.Lookuper) (interface{}, error) {
	return "Linux", nil
}

func (o OSLookup) GetDeviceType(l *lookups.Lookuper) (interface{}, error) {
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

func (o OSLookup) GetOSFullName(l *lookups.Lookuper) (interface{}, error) {
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
		if o.detectPro(l) {
			fullName = util.MakeNamePro(fullName)
		}
	}
	return fullName, nil
}

// detectPro attempts to determine if the ESM repos indicate that this server is
// getting full 'Pro' support
func (o OSLookup) detectPro(l *lookups.Lookuper) bool {
	out, err := l.Commander.Output("/usr/bin/pro", "security-status", "--esm-infra", "--format", "json")
	if err != nil {
		// No pro yo
		return false
	}
	var esm esmStatus
	err = json.Unmarshal(out, &esm)
	if err != nil {
		// Unknown pro yo
		return false
	}
	return esm.Summary.UA.Attached
}

// esmStatus is a minimal holder for the esm status
type esmStatus struct {
	Summary struct {
		UA struct {
			Attached bool `json:"attached"`
		} `json:"ua"`
	} `json:"summary"`
}

func (o OSLookup) GetInstalledSoftware(l *lookups.Lookuper) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (o OSLookup) GetHostname(_ *lookups.Lookuper) (interface{}, error) {
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return hostname, nil
}

func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookuper) (interface{}, error) {
	ids := map[string]string{}
	aidOut, err := l.Commander.Output("/opt/CrowdStrike/falconctl", "-g", "--aid")
	if err != nil {
		return nil, err
	}
	ids["crowdstrike_aid"] = strings.TrimSuffix(strings.TrimPrefix(string(aidOut), "aid=\""), "\".\n")
	return ids, nil
}
