package darwin_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/os/darwin"
)

type MockCommander struct{}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

// mock cmd.Execute
func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	joined := fmt.Sprintf("%v %v", command, strings.Join(args, " "))
	switch joined {
	case "/usr/sbin/scutil --get LocalHostName":
		return []byte("test-hostname"), nil
	case "/usr/sbin/sysctl -n hw.memsize":
		return []byte("8388608"), nil
	case "/usr/local/bin/brew list --versions":
		return []byte(`unixodbc 2.3.9_1
utf8proc 2.6.1
utimer 0.4_1`), nil
	case "/usr/sbin/ioreg -rd1 -c IOPlatformExpertDevice":
		return []byte(`+-o MacBookPro16,1  <class IOPlatformExpertDevice, id 0x100000114, registered, matched, active, busy 0 (67464 ms), retain 53>
		{
		  "IOInterruptSpecifiers" = (<0900000005000000>)
		  "IOPolledInterface" = "SMCPolledInterface is not serializable"
		  "IOPlatformUUID" = "15A89B8F-8BE9-4E84-B425-A8D35D6A3D7B"
		  "serial-number" = <4d443657000000000000000000000000000000000000000000000000000000>
		  "platform-feature" = <3200000000000000>
		  "IOPlatformSystemSleepPolicy" = <0000000000000a0008000000080000000000000000000000050000000000000005010000010000000000400000004000000010000000100007000000000000000fdd8701000000002000000020000000000000000000000005$
		  "IOBusyInterest" = "IOCommand is not serializable"
		  "target-type" = <"Mac">
		  "IOInterruptControllers" = ("io-apic-0")
		  "name" = <"/">
		  "version" = <"1.0">
		  "manufacturer" = <"Apple Inc.">
		  "compatible" = <"MacBookPro16,1">
		  "product-name" = <"MacBookPro16,1">
		  "IOPlatformSerialNumber" = "serial-yo"
		  "IOConsoleSecurityInterest" = "IOCommand is not serializable"
		  "clock-frequency" = <0084d717>
		  "model" = <"MacBookPro16,1">
		  "board-id" = <"Mac-someid">
		  "bridge-model" = <"J152fAP">
		  "IOGeneralInterest" = "IOCommand is not serializable"
		  "system-type" = <02>
		}
		`), nil

	case "/usr/sbin/system_profiler SPApplicationsDataType -json":
		return []byte(`{
			"SPApplicationsDataType" : [
			  {
			    "_name" : "Installer",
			    "arch_kind" : "arch_arm_i64",
			    "lastModified" : "2021-10-18T03:30:38Z",
			    "obtained_from" : "apple",
			    "path" : "/System/Library/CoreServices/Installer.app",
			    "signed_by" : [
			      "Software Signing",
			      "Apple Code Signing Certification Authority",
			      "Apple Root CA"
			    ],
			    "version" : "6.2.0"
			  },
			  {
			    "_name" : "TextEdit",
			    "arch_kind" : "arch_arm_i64",
			    "lastModified" : "2021-10-18T03:30:38Z",
			    "obtained_from" : "apple",
			    "path" : "/System/Applications/TextEdit.app",
			    "signed_by" : [
			      "Software Signing",
			      "Apple Code Signing Certification Authority",
			      "Apple Root CA"
			    ],
			    "version" : "1.17"
			  }]}`), nil

	case "/usr/sbin/system_profiler SPSoftwareDataType -json":
		return []byte(`{
			"SPSoftwareDataType" : [
			  {
			    "_name" : "os_overview",
			    "boot_mode" : "normal_boot",
			    "boot_volume" : "Macintosh HD",
			    "kernel_version" : "Darwin 21.1.0",
			    "local_host_name" : "test-hostname",
			    "os_version" : "macOS 12.0.1 (21A559)",
			    "secure_vm" : "secure_vm_enabled",
			    "system_integrity" : "integrity_enabled",
			    "uptime" : "up 0:16:15:15",
			    "user_name" : "Joe User (joeu)"
			  }
			]
		      }`), nil
	default:
		return []byte("fell through to default"), nil
	}
}

// This is the worlds worst command runner, always returns an error
type ErrorMockCommander struct{}

func (c ErrorMockCommander) Output(command string, args ...string) ([]byte, error) {
	return nil, errors.New("Always errors")
}

var commander cmdr.Commander

func TestGetHostname(t *testing.T) {
	commander = MockCommander{}
	c := lookups.LookuperConfig{
		Commander: &commander,
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
	commander = MockCommander{}
	c := lookups.LookuperConfig{
		Commander: &commander,
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
}
