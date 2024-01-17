package linux_test

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"

	report "gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

var (
	// Stubs for centos8
	centos8Data *report.MockData
	commander   report.Commander

	// Stubs for raspberry pi
	piData      *report.MockData
	piCommander report.Commander

	// Worlds worst command executor, always fail
	failCommander report.Commander
)

func setup() {
	var err error
	centos8Data, err = report.NewMockData("testdata/centos8.yaml")
	if err != nil {
		slog.Error("error setting centos8 data", "error", err)
		os.Exit(2)
	}
	piData, err = report.NewMockData("testdata/raspberrypi.yaml")
	if err != nil {
		slog.Error("error setting pi data", "error", err)
		os.Exit(2)
	}
}

func teardown() {
}

type (
	MockCommander struct{}
	PiCommander   struct{}
	FailCommander struct{}
)

// Mock for Centos8 good stuff
func (c MockCommander) Slurp(filepath string) ([]byte, error) {
	return report.MockFileGet(centos8Data, filepath)
}

func (c MockCommander) GetMacAddrs() ([]string, error) {
	return report.MockMacGet(centos8Data)
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	return report.MockCommandGet(centos8Data, command, args...)
}

// Mock up for raspberry pi
func (c PiCommander) GetMacAddrs() ([]string, error) {
	return report.MockMacGet(piData)
}

func (c PiCommander) LookPath(command string) (string, error) {
	availableCommands := []string{"dpkg-query"}
	if slices.Contains(availableCommands, command) {
		return fmt.Sprintf("/usr/bin/%v", command), nil
	}
	return "", errors.New("Command not found")
}

func (c PiCommander) Output(command string, args ...string) ([]byte, error) {
	return report.MockCommandGet(piData, command, args...)
}

func (c PiCommander) Slurp(filepath string) ([]byte, error) {
	return report.MockFileGet(piData, filepath)
}

// Mock up for failure cmds
func (c FailCommander) GetMacAddrs() ([]string, error) {
	return report.MockMacGet(nil)
}

func (c FailCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c FailCommander) Output(_ string, _ ...string) ([]byte, error) {
	return nil, errors.New("Always-fail")
}

func (c FailCommander) Slurp(filepath string) ([]byte, error) {
	return report.MockFileGet(nil, filepath)
}
