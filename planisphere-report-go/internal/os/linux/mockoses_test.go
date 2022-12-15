package linux_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/testlib"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/util"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

var (
	// Stubs for centos8
	centos8Data *testlib.MockData
	commander   cmdr.Commander

	// Stubs for raspberry pi
	piData      *testlib.MockData
	piCommander cmdr.Commander

	// Worlds worst command executor, always fail
	failCommander cmdr.Commander
)

func setup() {
	var err error
	centos8Data, err = testlib.NewMockData("testdata/centos8.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting centos8 data")
	}
	piData, err = testlib.NewMockData("testdata/raspberrypi.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting pi data")
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
	res, err := testlib.MockFileGet(centos8Data, filepath)
	return res, err
}

func (c MockCommander) GetMacAddrs() ([]string, error) {
	addrs, err := testlib.MockMacGet(centos8Data)
	return addrs, err
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	res, err := testlib.MockCommandGet(centos8Data, command, args...)
	return res, err
}

// Mock up for raspberry pi
func (c PiCommander) GetMacAddrs() ([]string, error) {
	addrs, err := testlib.MockMacGet(piData)
	return addrs, err
}

func (c PiCommander) LookPath(command string) (string, error) {
	availableCommands := []string{"dpkg-query"}
	if util.ContainsString(availableCommands, command) {
		return fmt.Sprintf("/usr/bin/%v", command), nil
	}
	return "", errors.New("Command not found")
}

func (c PiCommander) Output(command string, args ...string) ([]byte, error) {
	res, err := testlib.MockCommandGet(piData, command, args...)
	return res, err
}

func (c PiCommander) Slurp(filepath string) ([]byte, error) {
	res, err := testlib.MockFileGet(piData, filepath)
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
	// res, err := testlib.MockCommandGet(piData, command, args...)
	return nil, errors.New("Always-fail")
}

func (c FailCommander) Slurp(filepath string) ([]byte, error) {
	res, err := testlib.MockFileGet(nil, filepath)
	return res, err
}
