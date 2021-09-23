//go:build freebsd
// +build freebsd

package helpers

import (
	"errors"
)

func ApplyPlatformDetections(l *Lookuper) error {
	// Do initializing bits here
	return nil
}
func setSerial(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setManufacturer(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setModel(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
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
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setDeviceType(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setOSFullName(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setUsername(l *Lookuper) (string, error) {
	// TODO: Implement this
	return "", errors.New("Not yet implemented")
}

func setInstalledSoftware(l *Lookuper) ([][]string, error) {
	// TODO: Implement this
	return nil, errors.New("Not yet implemented")

}
