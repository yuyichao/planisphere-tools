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
	// TODO: Implement this
	return "", errors.New("Serial not yet implemented")
}

func setManufacturer(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	return "", errors.New("Manufacturer not yet implemented")
}

func setModel(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	return "", errors.New("Model not yet implemented")
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
	// TODO: Implement this
	return "", errors.New("DeviceType not yet implemented")
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
