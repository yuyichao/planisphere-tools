package cmdr

import "os/exec"

type Commander interface {
	Output(string, ...string) ([]byte, error)
	LookPath(string) (string, error)
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
