//go:build freebsd
// +build freebsd

package helpers

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ApplyPlatformDetections(l *Lookuper) error {
	// Do initializing bits here
	return nil
}
func setSerial(l *Lookuper) (interface{}, error) {
	cmdPath, err := exec.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up serial number")
	}
	cmdOut, err := exec.Command(cmdPath, "-s", "system-serial-number").Output()
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the serial number")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func setManufacturer(l *Lookuper) (interface{}, error) {
	cmdPath, err := exec.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up manufacturer")
	}
	cmdOut, err := exec.Command(cmdPath, "-s", "chassis-manufacturer").Output()
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the chassis-manufacturer")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func setModel(l *Lookuper) (interface{}, error) {
	cmdPath, err := exec.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up model")
	}
	cmdOut, err := exec.Command(cmdPath, "-s", "chassis-version").Output()
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the chassis-version")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func setDiskEncrypted(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	return false, errors.New("DiskEcrypted not yet implemented")
}

func setMemory(l *Lookuper) (interface{}, error) {
	out, err := exec.Command("/sbin/sysctl", "-n", "vm.kmem_size").Output()
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

func setOSFamily(l *Lookuper) (interface{}, error) {
	return "FreeBSD", nil
}

func setDeviceType(l *Lookuper) (interface{}, error) {
	cmdPath, err := exec.LookPath("dmidecode")
	if err != nil {
		return "", errors.New("dmidecode is needed to look up model")
	}
	cmdOut, err := exec.Command(cmdPath, "-s", "chassis-type").Output()
	if err != nil {
		return "", errors.New("Issue running dmidecode to get the chassis-type")
	}
	trimmed := strings.Trim(string(cmdOut), "\n")
	return trimmed, nil
}

func setOSFullName(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	out, err := exec.Command("/bin/freebsd-version").Output()
	if err != nil {
		return "", errors.New("Could not run freebsd-version successfully")
	}
	trimmed := strings.Trim(string(out), "\n")
	return trimmed, nil
}

func setUsername(l *Lookuper) (interface{}, error) {
	out, err := exec.Command("/usr/bin/last").Output()
	if err != nil {
		log.Warning("Error running 'last' to determine the real user")
	}
	ignoreUsers := []string{"root", "shutdown", ""}
	for _, line := range strings.Split(string(out), "\n") {
		pieces := strings.Split(line, " ")
		piece := pieces[0]
		if !ContainsString(ignoreUsers, piece) {
			return piece, nil
		}
	}
	return "", errors.New("Could not find a non-root user")
}

type BSDSoftware struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func GetInstalledSoftware() ([][]string, error) {
	softwareTable := [][]string{}

	cmdPath, err := exec.LookPath("pkg")
	if err == nil {
		cmdOut, err := exec.Command(cmdPath, "info", "--raw", "-a", "--raw-format", "json-compact").Output()
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

func setInstalledSoftware(l *Lookuper) (interface{}, error) {
	apps, err := GetInstalledSoftware()
	if err != nil {
		return nil, err
	}
	return apps, nil

}

func setHostname(l *Lookuper) (interface{}, error) {
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err

	}
	return hostname, nil
}
