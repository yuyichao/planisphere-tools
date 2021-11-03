package testlib

import "errors"

/*
Unsure if this will be usable. Trying to get a good way to move the mock command
and file output out of the test files, without creating waaaaay too many
testdata files to read in
*/

type MockData struct {
	Name          string            `yaml:"name,omitempty"`
	CommandOutput map[string]string `yaml:"command_output,omitempty"`
	FileContents  map[string]string `yaml:"file_contents,omitempty"`
}

func MockCommandGet(d *MockData, cmd string) (string, error) {
	if val, ok := d.CommandOutput[cmd]; ok {
		return val, nil
	}
	return "unknown-command", errors.New("unknown-mock-command")
}

func MockFileGet(d *MockData, fp string) (string, error) {
	if val, ok := d.FileContents[fp]; ok {
		return val, nil
	}
	return "unknown-file", errors.New("unknown-mock-file")
}
