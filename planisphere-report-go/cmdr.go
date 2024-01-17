package report

import (
	"net"
	"os"
	"os/exec"
	"path"
)

// Commander is an interface to do OS type command stuff
type Commander interface {
	Output(string, ...string) ([]byte, error)
	LookPath(string) (string, error)
	GetMacAddrs() ([]string, error)
	Slurp(string) ([]byte, error)
}

// RealCommander is the default implementation of a Commander
type RealCommander struct{}

// Output outputs exec strings
func (c RealCommander) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

// LookPath looks up the full path of an executable
func (c RealCommander) LookPath(command string) (string, error) {
	p, err := exec.LookPath(command)
	return p, err
}

// GetMacAddrs returns mac addresses
func (c RealCommander) GetMacAddrs() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}

// Slurp reads a file in and returns the byte content
func (c RealCommander) Slurp(filepath string) ([]byte, error) {
	b, err := os.ReadFile(path.Clean(filepath))
	if err != nil {
		return nil, err
	}

	return b, err
}
