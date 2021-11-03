package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

func TestNewLookuper(t *testing.T) {
	overrides := map[string]interface{}{}
	c := &lookups.LookuperConfig{
		Overrides: overrides,
	}
	_, err := helpers.NewLookuper(c)
	require.NoError(t, err)
}

func TestGenericLookupOverrides(t *testing.T) {
	overrides := map[string]interface{}{
		"instance_key":       "ringo",
		"department_key":     "foo",
		"support_group_id":   42,
		"support_group_name": "Marty",
		"usage_type":         "server",
		"status":             "deployed",
		"hostname":           "sumhost.local",
	}
	c := &lookups.LookuperConfig{
		Overrides: overrides,
	}
	l, _ := helpers.NewLookuper(c)
	require.Equal(t, "ringo", l.Payload.Key)
	require.Equal(t, "foo", l.Payload.Data.DepartmentKey)
	require.Equal(t, int(42), int(l.Payload.Data.SupportGroupId))
	require.Equal(t, "Marty", l.Payload.Data.SupportGroupName)
	require.Equal(t, "server", l.Payload.Data.UsageType)
	require.Equal(t, "deployed", l.Payload.Data.Status)
	require.Equal(t, "sumhost.local", l.Payload.Data.Hostname)
}

func TestOSSpecificLookupOverrides(t *testing.T) {
	overrides := map[string]interface{}{
		"manufacturer":   "ACME Inc.",
		"disk_encrypted": true,
		"memory_mb":      500,
		"serial":         "ABC-123",
		"os_family":      "Toretto",
		"model":          "Bruno",
	}
	c := &lookups.LookuperConfig{
		Overrides: overrides,
	}
	l, _ := helpers.NewLookuper(c)

	require.Equal(t, "ACME Inc.", l.Payload.Data.Manufacturer)
	require.True(t, l.Payload.Data.DiskEncrypted)
	require.Equal(t, int(500), int(l.Payload.Data.MemoryMB))
	require.Equal(t, "ABC-123", l.Payload.Data.Serial)
	require.Equal(t, "Toretto", l.Payload.Data.OsFamily)
	require.Equal(t, "Bruno", l.Payload.Data.Model)
}
