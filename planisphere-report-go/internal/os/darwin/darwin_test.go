package darwin_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/darwin"
)

func TestGetHostname(t *testing.T) {
	macOSCommander = MockCommander{}
	c := lookups.LookupConfig{
		Commander: &macOSCommander,
		OS:        "darwin",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)

	o := darwin.OSLookup{}
	hostname, err := o.GetHostname(l)
	require.NoError(t, err)
	require.Equal(t, "test-hostname", hostname)
}

func TestNewLookup(t *testing.T) {
	macOSCommander = MockCommander{}
	c := lookups.LookupConfig{
		Commander: &macOSCommander,
		OS:        "darwin",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "macOS", l.Payload.Data.OsFamily)
	require.Equal(t, "macOS 12.0.1 (21A559)", l.Payload.Data.OsFullname)
	require.Equal(t, int(8), int(l.Payload.Data.MemoryMB))
	require.Equal(t, "MacBookPro16,1", l.Payload.Data.Model)
	require.Equal(t, "laptop", l.Payload.Data.DeviceType)
	// How should we test this if we only run as root?
	// require.Equal(t, planisphere.ExternalOSIdentifiers{"crowdstrike_aid": "d3bcdcaf1604426b9ad6a421e1a5ad40"}, l.Payload.Data.ExternalOSIdentifiers)
}

func TestFailingLookup(t *testing.T) {
	failCommander = FailCommander{}
	c := lookups.LookupConfig{
		Commander: &failCommander,
		OS:        "darwin",
	}
	// Lookups must be able to exec on osx to run the ioreg tool
	_, err := helpers.NewLookuper(&c)
	require.Error(t, err)
}
