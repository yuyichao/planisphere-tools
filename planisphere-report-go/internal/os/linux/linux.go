package linux

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/hardware"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

var (
	ErrEmptyOutput        = fmt.Errorf("Empty output")
	ErrMalformedRPMOutput = fmt.Errorf("Malformed RPM Output")
)

type OSLookup struct{}

func (o OSLookup) ApplyPlatformDetections(l *lookups.Lookuper) error {
	return nil
}

func GetInstalledSoftware(l *lookups.Lookuper) ([][]string, error) {
	softwareTable := [][]string{}

	// RPMs
	rpmPath, err := l.Commander.LookPath("rpm")
	if err == nil {
		rpmOut, err := l.Commander.Output(rpmPath, "-qa", "--qf", "%{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n")
		if err != nil {
			log.Warning("Could not do an rpm listing even though the rpm command exists")
		}
		rpmSoftware, err := ParseRPMOutput(rpmOut)
		if err != nil {
			log.Warning("Could not parse the rpm output")
		} else {
			softwareTable = append(softwareTable, rpmSoftware...)
		}
	} else {
		log.Println("No rpm command installed")
	}

	// DPKG-Query
	cmdPath, err := l.Commander.LookPath("dpkg-query")
	if err == nil {
		cmdOut, err := l.Commander.Output(cmdPath, "-W")
		if err != nil {
			log.Warning("Could not do a dpkg-query listing even though the command exists")
		}
		trimmed := strings.Trim(string(cmdOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, "\t")
			if len(pieces) != 2 {
				log.Warningf("Got weird line from dpkg: %v", line)
			} else {
				name := pieces[0]
				version := pieces[1]

				softwareTable = append(softwareTable, []string{name, version})
			}
		}
	} else {
		log.Println("No dpkg-query command installed")
	}

	// Pacman nom nom nom
	cmdPath, err = l.Commander.LookPath("pacman")
	if err == nil {
		cmdOut, err := l.Commander.Output(cmdPath, "-Q")
		if err != nil {
			log.Warning("Could not do a pacman listing even though the command exists")
		}
		trimmed := strings.Trim(string(cmdOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, " ")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	} else {
		log.Println("No pacman command installed")
	}

	// Guix
	cmdPath, err = l.Commander.LookPath("guix-installed")
	if err == nil {
		cmdOut, err := l.Commander.Output(cmdPath)
		if err != nil {
			log.Warning("Could not do a guix-installed listing even though the command exists")
		}
		trimmed := strings.Trim(string(cmdOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, "\t")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	} else {
		log.Println("No guix-installed command installed")
	}

	return softwareTable, nil
}

func (o OSLookup) GetSerial(l *lookups.Lookuper) (interface{}, error) {
	out, err := l.Slurper.Slurp("/sys/class/dmi/id/product_serial")
	if err != nil {
		return nil, err
	}
	serial := strings.TrimSpace(string(out))
	return serial, nil
}

func (o OSLookup) GetManufacturer(l *lookups.Lookuper) (interface{}, error) {
	// Is it a Raspberry Pi?
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			return "Raspberry Pi", nil
		}
	}
	out, err := l.Slurper.Slurp("/sys/class/dmi/id/bios_vendor")
	if err != nil {
		return nil, err
	}
	vendor := strings.TrimSuffix(string(out), "\n")
	return vendor, nil
}

func (o OSLookup) GetModel(l *lookups.Lookuper) (interface{}, error) {
	// Is it a Raspberry Pi?
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			cpuDat, err := os.ReadFile("/proc/cpuinfo")
			if err != nil {
				log.Warning(err)
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
	out, err := l.Slurper.Slurp("/sys/class/dmi/id/product_name")
	if err != nil {
		return nil, err
	}
	productName := strings.TrimSuffix(string(out), "\n")
	return productName, nil
}

func (o OSLookup) GetDiskEncrypted(l *lookups.Lookuper) (interface{}, error) {
	// TODO: Implement this
	return false, errors.New("DiskEncrypted Not yet implemented")
}

func (o OSLookup) GetMemory(l *lookups.Lookuper) (interface{}, error) {
	// TODO: Implement this
	return 0, errors.New("Memory yet implemented")
}

func (o OSLookup) GetOSFamily(l *lookups.Lookuper) (interface{}, error) {
	return "Linux", nil
}

func (o OSLookup) GetDeviceType(l *lookups.Lookuper) (interface{}, error) {
	out, err := l.Slurper.Slurp("/sys/class/dmi/id/chassis_type")
	if err != nil {
		return nil, err
	}
	cid, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return nil, err
	}
	var chassisType string
	if val, ok := hardware.ChassisType[uint(cid)]; ok {
		chassisType = val
	} else {
		chassisType = "Other"
	}
	return chassisType, nil
}

func (o OSLookup) GetOSFullName(l *lookups.Lookuper) (interface{}, error) {
	osb, err := l.Slurper.Slurp("/etc/os-release")
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
	return fullName, nil
}

func (o OSLookup) GetInstalledSoftware(l *lookups.Lookuper) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (o OSLookup) GetHostname(l *lookups.Lookuper) (interface{}, error) {
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return hostname, nil
}
