package freebsd_test

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
	// MockSlurper   struct{}
)

var (
	commander cmdr.Commander
	slurper   cmdr.Slurper
)

func (c MockCommander) Output(command string, args ...string) ([]byte, error) {
	joined := fmt.Sprintf("%v %v", command, strings.Join(args, " "))
	joined = strings.TrimSpace(joined)
	switch joined {
	case "/usr/bin/dmidecode -s chassis-type":
		return []byte("Other"), nil
	case "/usr/bin/dmidecode -s system-serial-number":
		return []byte("some-fake-serial-number"), nil
	case "dmidecode -s chassis-version":
		return []byte("N/A"), nil
	case "/usr/bin/dmidecode -s chassis-manufacturer":
		return []byte("No Enclosure"), nil
	case "/usr/bin/pkg info --raw -a --raw-format json-compact":
		return []byte(`{"name":"wireguard-go","origin":"net/wireguard-go","version":"0.0.20210424,1","comment":"WireGuard implementation in Go","maintainer":"decke@FreeBSD.org","www":"https://www.wireguard.com","abi":"FreeBSD:12:amd64","arch":"freebsd:12:x86:64","prefix":"/usr/local","flatsize":2684290,"timestamp":1628675118,"licenselogic":"single","licenses":["MIT"],"desc":"This is an implementation of Wireguard in Go.\n\nWireGuard is an extremely simple yet fast and modern VPN that utilizes\nstate-of-the-art cryptography. It aims to be faster, simpler, leaner,\nand more useful than IPSec, while avoiding the massive headache. It\nintends to be considerably more performant than OpenVPN. WireGuard is\ndesigned as a general purpose VPN for running on embedded interfaces and\nsuper computers alike, fit for many different circumstances.\n\nWWW: https://www.wireguard.com","categories":["net-vpn","net"],"annotations":{"FreeBSD_version":"1201000","repo_type":"binary","repository":"OPNsense"},"files":{"/usr/local/bin/wireguard-go":"1$fd335ae0c2c7c5d9e9c618e6ff33e868c61be2269238792e3a55ebe25286c033","/usr/local/share/licenses/wireguard-go-0.0.20210424,1/LICENSE":"1$8a9617637463b68fc0584861100926c56e2af396c013024314c285549b9c9642","/usr/local/share/licenses/wireguard-go-0.0.20210424,1/MIT":"1$91276db973f25602d1aa43491f59cbc84cb88e6f151e1d0cc82a755563ce0195","/usr/local/share/licenses/wireguard-go-0.0.20210424,1/catalog.mk":"1$b18631491780dfbaf8904fc5182ea18f1e48831ee95402eb22358fada74d99b1"}}
{"name":"zip","origin":"archivers/zip","version":"3.0_1","comment":"Create/update ZIP files compatible with PKZIP","maintainer":"ler@FreeBSD.org","www":"http://infozip.sourceforge.net/Zip.html","abi":"FreeBSD:12:amd64","arch":"freebsd:12:x86:64","prefix":"/usr/local","flatsize":532422,"timestamp":1628675117,"licenselogic":"single","licenses":["BSD3CLAUSE"],"desc":"Zip is a compression and file packaging utility.  It is compatible with\nPKZIP 2.04g (Phil Katz ZIP) for MSDOS systems.  There is a companion to zip\ncalled unzip (of course) which you can also install from the ports/package\nsystem.\n\nWWW: http://infozip.sourceforge.net/Zip.html","categories":["archivers"],"options":{"DOCS":"off"},"annotations":{"FreeBSD_version":"1201000","repo_type":"binary","repository":"OPNsense"},"files":{"/usr/local/bin/zip":"1$1dc2866fcc35b80410525c1dbdeabb3576b42ecd359148451cffa78beb35673a","/usr/local/bin/zipcloak":"1$1c941359d3fa671d61b08b959cdfe2abc6a80d1c7fd9f012fbb83bb3c5e0b454","/usr/local/bin/zipnote":"1$3be27b94c0a820ba794dbee6db36fb429699ad81c426af4d3bc5dad029efb6bd","/usr/local/bin/zipsplit":"1$1169688cdcfa3ce690c09e680d61acd5f3ccf36cf578a7e1f4cbbe2e866fc3c6","/usr/local/man/man1/zip.1.gz":"1$9f984b4b87db8f5ccf78dd28c9907a601f8c0a431d4efd64195b5006cf189818","/usr/local/man/man1/zipcloak.1.gz":"1$bcf90b8d44d97147ed1193a5c1ad7938b0dda636989b251264967c5de3100854","/usr/local/man/man1/zipnote.1.gz":"1$97ad30a152d4536a873116c13be3bbf901778ba6f194914f14b2b29b4c24c252","/usr/local/man/man1/zipsplit.1.gz":"1$c72d170ffee8346bf09c01ee79c99b108188d7f7420a4560a162e0e82d5a8bff","/usr/local/share/licenses/zip-3.0_1/BSD3CLAUSE":"1$8ecd6c1bab449127eb665cef1561e73a8bce52e217375f6f466939e137b1e110","/usr/local/share/licenses/zip-3.0_1/LICENSE":"1$dba31a5da2c9507fa0fc58af4447cb05d27ecac7e99b9cd195cbca1605bf7bb0","/usr/local/share/licenses/zip-3.0_1/catalog.mk":"1$7f6e0b553b537740a810de629d871ec8b45f2e39acb285b1bfbf61ea2b0cbcf4"}}`), nil
	case "/bin/freebsd-version":
		return []byte("FreeBSD 11.1-RELEASE-p4 amd64"), nil
	case "/sbin/sysctl -n vm.kmem_size":
		return []byte("8388608"), nil
	default:
		log.Warningf("Unknown command: %v", joined)
		return []byte("fell through"), nil
	}
}

func (c MockCommander) LookPath(command string) (string, error) {
	return fmt.Sprintf("/usr/bin/%v", command), nil
}

func TestNewLookup(t *testing.T) {
	commander = MockCommander{}
	// slurper = MockSlurper{}
	c := lookups.LookuperConfig{
		Commander: &commander,
		OS:        "freebsd",
		// Slurper:   &slurper,
	}
	l, err := helpers.NewLookuper(&c)
	require.NoError(t, err)
	require.NotNil(t, l)

	require.Equal(t, "FreeBSD", l.Payload.Data.OsFamily)
	require.Equal(t, "FreeBSD 11.1-RELEASE-p4 amd64", l.Payload.Data.OsFullname)
	require.Equal(t, "N/A", l.Payload.Data.Model)
	require.Equal(t, "Other", l.Payload.Data.DeviceType)
}
