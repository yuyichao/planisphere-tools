package helpers_test

import (
	"testing"

	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
)

func TestNewLookuper(t *testing.T) {
	overrides := map[string]interface{}{}
	_, err := helpers.NewLookuper(overrides)
	if err != nil {
		t.Error("Could successfully create a NewLookuper")
	}

}
