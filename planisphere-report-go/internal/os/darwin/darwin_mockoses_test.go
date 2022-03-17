package darwin_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/apex/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/testlib"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

var (
	// Stubs for centos8
	macOSData      *testlib.MockData
	macOSCommander cmdr.Commander

	// Worlds worst command executor, always fail
	failCommander cmdr.Commander
)

func setup() {
	var err error
	macOSData, err = testlib.NewMockData("testdata/macos.yaml")
	if err != nil {
		log.WithError(err).Fatal("Error setting macOS data")
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
	res, err := testlib.MockFileGet(macOSData, filepath)
	return res, err
}

func (c MockCommander) GetMacAddrs() ([]string, error) {
	addrs, err := testlib.MockMacGet(macOSData)
	return addrs, err
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	res, err := testlib.MockCommandGet(macOSData, command, args...)
	return res, err
}

// Mock up for failure cmds
func (c FailCommander) GetMacAddrs() ([]string, error) {
	addrs, err := testlib.MockMacGet(nil)
	return addrs, err
}

func (c FailCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c FailCommander) Output(command string, args ...string) ([]byte, error) {
	return nil, errors.New("Always-fail")
}

func (c FailCommander) Slurp(filepath string) ([]byte, error) {
	res, err := testlib.MockFileGet(nil, filepath)
	return res, err
}
