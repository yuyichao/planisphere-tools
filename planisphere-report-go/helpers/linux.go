//go:build linux
// +build linux

package helpers

import (
	"errors"
	"os/exec"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/zcalusic/sysinfo"
)

var si sysinfo.SysInfo

func ApplyPlatformDetections(l *Lookuper) error {
	si.GetSysInfo()
	return nil
}

func GetInstalledSoftware() ([][]string, error) {
	softwareTable := [][]string{}

	// RPMs
	rpmPath, err := exec.LookPath("rpm")
	if err == nil {
		rpmOut, err := exec.Command(rpmPath, "-qa", "--qf", "%{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n").Output()
		if err != nil {
			log.Warning("Could not do an rpm listing even though the rpm command exists")
		}
		trimmed := strings.Trim(string(rpmOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, " ")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	} else {
		log.Println("No rpm command installed")
	}

	// DPKG-Query
	cmdPath, err := exec.LookPath("dpkg-query")
	if err == nil {
		cmdOut, err := exec.Command(cmdPath, "-W").Output()
		if err != nil {
			log.Warning("Could not do a dpkg-query listing even though the command exists")
		}
		trimmed := strings.Trim(string(cmdOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, " ")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	} else {
		log.Println("No dpkg-query command installed")
	}

	// Pacman nom nom nom
	cmdPath, err = exec.LookPath("pacman")
	if err == nil {
		cmdOut, err := exec.Command(cmdPath, "-Q").Output()
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
	cmdPath, err = exec.LookPath("guix-installed")
	if err == nil {
		cmdOut, err := exec.Command(cmdPath).Output()
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

func setSerial(l *Lookuper) (string, error) {
	return si.Product.Serial, nil
}

func setManufacturer(l *Lookuper) (string, error) {
	return si.Product.Vendor, nil
}

func setModel(l *Lookuper) (string, error) {
	return si.Product.Name, nil
}

func setDiskEncrypted(l *Lookuper) (bool, error) {
	// TODO: Implement this
	return false, errors.New("Not yet implemented")
}

func setMemory(l *Lookuper) (uint64, error) {
	// TODO: Implement this
	return 0, errors.New("Not yet implemented")
}

func setOSFamily(l *Lookuper) (string, error) {
	return "Linux", nil
}

func setDeviceType(l *Lookuper) (string, error) {
	if _, ok := ChassisType[si.Chassis.Type]; ok {
		return ChassisType[si.Chassis.Type], nil
	} else {
		return "", errors.New("Unknown Device Type")
	}
}

func setOSFullName(l *Lookuper) (string, error) {
	return si.OS.Name, nil
}

func setUsername(l *Lookuper) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

func setInstalledSoftware(l *Lookuper) ([][]string, error) {
	apps, err := GetInstalledSoftware()
	if err != nil {
		return nil, err
	}
	return apps, nil

}
