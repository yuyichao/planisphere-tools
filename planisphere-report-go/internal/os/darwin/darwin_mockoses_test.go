package darwin_test

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
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
	macOSData      *report.MockData
	macOSCommander report.Commander

	// Worlds worst command executor, always fail
	failCommander report.Commander
)

func setup() {
	var err error
	macOSData, err = report.NewMockData("testdata/macos.yaml")
	if err != nil {
		slog.Error("error setting macOS data", "error", err)
		os.Exit(2)
	}
}

func teardown() {
}

type (
	MockCommander struct{}
	FailCommander struct{}
)

// Mock for Centos8 good stuff
func (c MockCommander) Slurp(filepath string) ([]byte, error) {
	res, err := report.MockFileGet(macOSData, filepath)
	return res, err
}

func (c MockCommander) GetMacAddrs() ([]string, error) {
	addrs, err := report.MockMacGet(macOSData)
	return addrs, err
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	res, err := report.MockCommandGet(macOSData, command, args...)
	return res, err
}

// Mock up for failure cmds
func (c FailCommander) GetMacAddrs() ([]string, error) {
	addrs, err := report.MockMacGet(nil)
	return addrs, err
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
