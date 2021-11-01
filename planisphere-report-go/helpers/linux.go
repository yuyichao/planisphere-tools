//go:build linux
// +build linux

package helpers

import (
	"errors"
	"os"
	"os/exec"
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
	cmdPath, err := exec.LookPath("dpkg-query")
	if err == nil {
		cmdOut, err := exec.Command(cmdPath, "-W").Output()
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

func setSerial(l *Lookuper) (interface{}, error) {
	return si.Product.Serial, nil
}

func setManufacturer(l *Lookuper) (interface{}, error) {
	// Is it a Raspberry Pi?
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			return "Raspberry Pi", nil
		}
	}
	return si.Product.Vendor, nil
}

func setModel(l *Lookuper) (interface{}, error) {
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
					if _, ok := RaspberryPiModels[value]; ok {
						return RaspberryPiModels[value], nil
					}
				}
			}
		}
	}
	return si.Product.Name, nil
}

func setDiskEncrypted(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	return false, errors.New("DiskEncrypted Not yet implemented")
}

func setMemory(l *Lookuper) (interface{}, error) {
	// TODO: Implement this
	return 0, errors.New("Memory yet implemented")
}

func setOSFamily(l *Lookuper) (interface{}, error) {
	return "Linux", nil
}

func setDeviceType(l *Lookuper) (interface{}, error) {
	if _, ok := ChassisType[si.Chassis.Type]; ok {
		return ChassisType[si.Chassis.Type], nil
	}
	return "", errors.New("Unknown Device Type")
}

func setOSFullName(l *Lookuper) (interface{}, error) {
	return si.OS.Name, nil
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
