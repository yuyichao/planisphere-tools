package cmdr

import (
	"net"
	"os"
	"os/exec"
)

type Commander interface {
	Output(string, ...string) ([]byte, error)
	LookPath(string) (string, error)
	GetMacAddrs() ([]string, error)
	Slurp(string) ([]byte, error)
}

type RealCommander struct{}

// mock cmd.Execute
func (c RealCommander) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

func (c RealCommander) LookPath(command string) (string, error) {
	p, err := exec.LookPath(command)
	return p, err
}

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

func (c RealCommander) Slurp(filepath string) ([]byte, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return b, err
}
