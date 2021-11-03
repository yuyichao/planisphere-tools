package linux

import "strings"

func ParseRPMOutput(rpmOut []byte) ([][]string, error) {
	softwareTable := [][]string{}
	lines := []string{}
	trimmed := strings.Trim(string(rpmOut), "\n")
	// RPM can run successfully, but output nothing. Handle this by checking for a new line for now. See:
	// https://gitlab.oit.duke.edu/devil-ops/planisphere-tools/-/issues/3 for more info
	if strings.Contains(trimmed, "\n") {
		// Consider empty if just a new line return
		lines = strings.Split(trimmed, "\n")
	} else {
		// Consider empty if completely blank
		if trimmed == "" {
			return nil, ErrEmptyOutput
		} else if trimmed == "\n" {
			return nil, ErrEmptyOutput
		}
		lines = append(lines, trimmed)
	}
	for _, line := range lines {
		pieces := strings.Split(line, " ")
		if len(pieces) != 2 {
			return nil, ErrMalformedRPMOutput
		}
		name := pieces[0]
		version := pieces[1]

		softwareTable = append(softwareTable, []string{name, version})
	}

	return softwareTable, nil
}
