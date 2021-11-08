package linux

import "strings"

func ParsePackageOutput(rpmOut []byte) ([][]string, error) {
	softwareTable := [][]string{}
	lines := []string{}
	trimmed := strings.Trim(string(rpmOut), "\n")
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
		pieces := strings.Fields(line)
		if len(pieces) != 2 {
			return nil, ErrMalformedPackageOutput
		}
		name := pieces[0]
		version := pieces[1]

		softwareTable = append(softwareTable, []string{name, version})
	}

	return softwareTable, nil
}
