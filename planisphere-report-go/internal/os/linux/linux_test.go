package linux_test

import (
	"fmt"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/cmdr"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

type (
	MockCommander struct{}
	MockSlurper   struct{}
)

func (c MockSlurper) Slurp(filepath string) ([]byte, error) {
	switch filepath {
	case "/sys/class/dmi/id/product_serial":
		return []byte("my-awesome-fake-serial"), nil
	case "/sys/class/dmi/id/bios_vendor":
		return []byte("VMware, Inc."), nil
	case "/sys/class/dmi/id/product_name":
		return []byte("VMware7,1"), nil
	case "/sys/class/dmi/id/chassis_type":
		return []byte("1"), nil
	case "/etc/os-release":
		return []byte(`NAME="CentOS Stream"
VERSION="8"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="8"
PLATFORM_ID="platform:el8"
PRETTY_NAME="CentOS Stream 8"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:8"
HOME_URL="https://centos.org/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"
REDHAT_SUPPORT_PRODUCT="Red Hat Enterprise Linux 8"
REDHAT_SUPPORT_PRODUCT_VERSION="CentOS Stream"`), nil
	default:
		log.Warning("Unknown path mock-slurped:", filepath)
		return nil, fmt.Errorf("Unknown path mock-slurped: %v", filepath)
	}
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

// mock cmd.Execute
func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	joined := fmt.Sprintf("%v %v", command, strings.Join(args, " "))
	joined = strings.TrimSpace(joined)
	switch joined {
	case "/usr/bin/rpm -qa --qf %{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}":
		return []byte(`mokutil 1:0.3.0-11.el8
python3-certbot 1.20.0-1.el8
xfce4-session 4.16.0-3.el8`), nil
	case "/usr/bin/dpkg-query -W":
		return []byte(`vim-common	2:8.1.2269-1ubuntu5.3
vim-runtime	2:8.1.2269-1ubuntu5.3
vim-tiny	2:8.1.2269-1ubuntu5.3
wamerican	2018.04.16-1`), nil
	case "/usr/bin/guix-installed":
		return []byte("foo\t1.2.3\nbar\t4.5.6"), nil
	case "/usr/bin/pacman -Q":
		return []byte(`foo 1.2.3
bar 4.5.6`), nil
	default:
		log.Warningf("Unknown command: %v", joined)
		return []byte("fell through"), nil
	}
}

var (
	commander cmdr.Commander
	slurper   cmdr.Slurper
)

func TestNewLookup(t *testing.T) {
	commander = MockCommander{}
	slurper = MockSlurper{}
	c := lookups.LookuperConfig{
		Commander: &commander,
		Slurper:   &slurper,
		OS:        "linux",
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "Linux", l.Payload.Data.OsFamily)
	require.Equal(t, "CentOS Stream 8", l.Payload.Data.OsFullname)
	require.Equal(t, "VMware7,1", l.Payload.Data.Model)
	require.Equal(t, "Other", l.Payload.Data.DeviceType)
	require.Equal(t, "my-awesome-fake-serial", l.Payload.Data.Serial)
}
