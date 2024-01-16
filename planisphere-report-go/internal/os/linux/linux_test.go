package linux_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

func TestNewLookup(t *testing.T) {
	commander = MockCommander{}
	c := lookups.LookupConfig{
		Commander: &commander,
		OS:        "linux",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "Linux", l.Payload.Data.OsFamily)
	require.Equal(t, "CentOS Stream 8", l.Payload.Data.OsFullname)
	require.Equal(t, "VMware7,1", l.Payload.Data.Model)
	require.Equal(t, "laptop", l.Payload.Data.DeviceType)
	require.Equal(t, "my-awesome-fake-serial", l.Payload.Data.Serial)
	require.Equal(t, planisphere.ExternalOSIdentifiers{"crowdstrike_aid": "d3bcdcaf1604426b9ad6a421e1a5ad40"}, l.Payload.Data.ExternalOSIdentifiers)
	require.Contains(t, l.Payload.Data.MacAddresses, "01:02:03:30:20:10")
}

func TestPiLookup(t *testing.T) {
	piCommander = PiCommander{}
	c := lookups.LookupConfig{
		Commander: &piCommander,
		OS:        "linux",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "Linux", l.Payload.Data.OsFamily)
	require.Equal(t, "Raspbian GNU/Linux 8 (jessie)", l.Payload.Data.OsFullname)
	require.Equal(t, "B", l.Payload.Data.Model)
	require.Equal(t, "", l.Payload.Data.DeviceType)
	require.Equal(t, "Raspberry Pi", l.Payload.Data.Manufacturer)
	require.Equal(t, "my-awesome-fake-serial", l.Payload.Data.Serial)
	require.Equal(t, int(31933), int(l.Payload.Data.MemoryMB))
	require.Contains(t, l.Payload.Data.MacAddresses, "b8:27:eb:30:20:10")
	// Make sure we detected some software
	require.Greater(t, len(l.Payload.Data.InstalledSoftware), 0, "No installed software found")
}

func TestFailingLookup(t *testing.T) {
	failCommander = FailCommander{}
	c := lookups.LookupConfig{
		Commander: &failCommander,
		OS:        "linux",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "Linux", l.Payload.Data.OsFamily)
	require.Equal(t, "", l.Payload.Data.OsFullname)
	require.Equal(t, "", l.Payload.Data.Model)
	require.Equal(t, "", l.Payload.Data.DeviceType)
	require.Equal(t, "", l.Payload.Data.Serial)
	require.Equal(t, planisphere.ExternalOSIdentifiers(nil), l.Payload.Data.ExternalOSIdentifiers)
	// require.Contains(t, l.Payload.Data.MacAddresses, "b8:27:eb:30:20:10")
}
