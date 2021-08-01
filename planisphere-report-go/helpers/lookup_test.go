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
	l, _ := helpers.NewLookuper(overrides)
	if l.Payload.Key != "ringo" {
		t.Error("Did not properly override InstanceKey")
	}

	if l.Payload.Data.DepartmentKey != "foo" {
		t.Error("Did not properly override DepartmentKey")
	}

	if l.Payload.Data.SupportGroupId != 42 {
		t.Error("Did not properly override SupportGroupId")
	}
	if l.Payload.Data.SupportGroupName != "Marty" {
		t.Error("Did not properly override SupportGroupName")
	}
	if l.Payload.Data.UsageType != "server" {
		t.Error("Did not properly override UsageType")
	}
	if l.Payload.Data.Status != "deployed" {
		t.Error("Did not properly override Status")
	}
	if l.Payload.Data.Hostname != "sumhost.local" {
		t.Error("Did not properly override Hostname")
	}
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
	l, _ := helpers.NewLookuper(overrides)

	if l.Payload.Data.Manufacturer != "ACME Inc." {
		t.Error("Did not properly override Manufactuer")
	}
	if !l.Payload.Data.DiskEncrypted {
		t.Error("Did not properly override Disk Encryption")
	}

	if l.Payload.Data.MemoryMB != 500 {
		t.Error("Did not properly override MemoryMB")
	}
	if l.Payload.Data.Serial != "ABC-123" {
		t.Error("Did not properly override Serial")
	}
	if l.Payload.Data.OsFamily != "Toretto" {
		t.Error("Did not properly override OSFamily")
	}
	if l.Payload.Data.Model != "Bruno" {
		t.Error("Did not properly override Model")
	}

}
