package freebsd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

var ErrDmidecode = fmt.Errorf("Error running dmidecode, ensure it is installed and you are running as root")

type OSLookup struct{}

func (o OSLookup) ApplyPlatformDetections(l *lookups.Lookuper) error {
	// Do initializing bits here
	return nil
}

func (o OSLookup) GetSerial(l *lookups.Lookuper) (interface{}, error) {
	cmdPath, err := l.Commander.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up serial number")
	}
	cmdOut, err := l.Commander.Output(cmdPath, "-s", "system-serial-number")
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the serial number")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func (o OSLookup) GetManufacturer(l *lookups.Lookuper) (interface{}, error) {
	cmdPath, err := l.Commander.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up manufacturer")
	}
	cmdOut, err := l.Commander.Output(cmdPath, "-s", "chassis-manufacturer")
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the chassis-manufacturer")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func (o OSLookup) GetModel(l *lookups.Lookuper) (interface{}, error) {
	out, err := l.Commander.Output("dmidecode", "-s", "chassis-version")
	if err != nil {
		return "", ErrDmidecode
	}
	trimmed := strings.Trim(string(out), "\n")
	return trimmed, nil
}

func (o OSLookup) GetDiskEncrypted(l *lookups.Lookuper) (interface{}, error) {
	// TODO: Implement this
	return false, errors.New("DiskEcrypted not yet implemented")
}

func (o OSLookup) GetMemory(l *lookups.Lookuper) (interface{}, error) {
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

func (o OSLookup) GetOSFamily(l *lookups.Lookuper) (interface{}, error) {
	return "FreeBSD", nil
}

func (o OSLookup) GetDeviceType(l *lookups.Lookuper) (interface{}, error) {
	l.WaitForChecked("serial")
	if strings.Contains(l.Payload.Data.Serial, "VMware") {
		return "vm", nil
	}
	return "", nil
}

func (o OSLookup) GetOSFullName(l *lookups.Lookuper) (interface{}, error) {
	// TODO: Implement this
	out, err := l.Commander.Output("/bin/freebsd-version")
	if err != nil {
		return "", errors.New("Could not run freebsd-version successfully")
	}
	trimmed := strings.Trim(string(out), "\n")
	return fmt.Sprintf("FreeBSD %v", trimmed), nil
}

type BSDSoftware struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func GetInstalledSoftware(l *lookups.Lookuper) ([][]string, error) {
	softwareTable := [][]string{}

	cmdPath, err := l.Commander.LookPath("pkg")
	if err == nil {
		cmdOut, err := l.Commander.Output(cmdPath, "info", "--raw", "-a", "--raw-format", "json-compact")
		if err != nil {
			log.Warning("Could not do a pkg info, even though the pkg command exists")
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
		log.Println("No pkg command installed")
	}

	return softwareTable, nil
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

func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookuper) (interface{}, error) {
	return nil, nil
}
