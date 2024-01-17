package report

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

/*
Unsure if this will be usable. Trying to get a good way to move the mock command
and file output out of the test files, without creating waaaaay too many
testdata files to read in
*/

// MockData is the data definition for a mock
type MockData struct {
	Name          string            `yaml:"name,omitempty"`
	CommandOutput map[string]string `yaml:"command_output,omitempty"`
	FileContents  map[string]string `yaml:"file_contents,omitempty"`
	MacAddresses  []string          `yaml:"mac_addresses,omitempty"`
}

// MockCommandGet yoinks a command out of the mock
func MockCommandGet(d *MockData, command string, args ...string) ([]byte, error) {
	joined := fmt.Sprintf("%v %v", command, strings.Join(args, " "))
	joined = strings.TrimSpace(joined)
	if val, ok := d.CommandOutput[joined]; ok {
		return []byte(val), nil
	}
	return []byte("unknown-command"), errors.New("unknown-mock-command")
}

// MockFileGet mocks the file get thing
func MockFileGet(d *MockData, fp string) ([]byte, error) {
	if d != nil {
		if val, ok := d.FileContents[fp]; ok {
			return []byte(val), nil
		}
		slog.Warn("missing mock file", "file", fp)
		return []byte("unknown-mock-file"), errors.New("unknown-mock-file")
	}
	return []byte(""), nil
}

// MockMacGet mocks the mac address getter
func MockMacGet(d *MockData) ([]string, error) {
	if d != nil {
		return d.MacAddresses, nil
	}
	return []string{"00:00:00:11:22:33"}, nil
}

// NewMockData returns a new MockData object
func NewMockData(filePath string) (*MockData, error) {
	var d MockData
	yamlFile, err := os.ReadFile(path.Clean(filePath))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
