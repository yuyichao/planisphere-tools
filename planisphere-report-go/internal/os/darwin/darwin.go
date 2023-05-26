package darwin

import (
	"encoding/json"
	"errors"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

type OSLookup struct{}

/*
Only platform specific stuff should be in here. If it's more generic than a
given platform, please include it in the NewLookup function
Is there anything we actually want to do here globally?
*/
// ApplyPlatformDetections Do some stuff here
var ioregExpert map[string]string

func (o OSLookup) GetDeviceType(l *lookups.Lookuper) (interface{}, error) {
	l.WaitForChecked("model")
	if strings.Contains(l.Payload.Data.Model, "MacBook") {
		return "laptop", nil
	}
	if strings.Contains(l.Payload.Data.Model, "Mac") {
		return "desktop", nil
	}
	return "", errors.New("Unknown type of mac")
}

func (o OSLookup) GetModel(_ *lookups.Lookuper) (interface{}, error) {
	return ioregExpert["model"], nil
}

func (o OSLookup) GetOSFullName(l *lookups.Lookuper) (interface{}, error) {
	softData, err := GetPSoftwareData(l)
	if err != nil {
		return "", err
	}

	return softData.SPSoftwareDataType[0].OsVersion, nil
}

func (o OSLookup) GetOSFamily(_ *lookups.Lookuper) (interface{}, error) {
	return "macOS", nil
}

func (o OSLookup) ApplyPlatformDetections(l *lookups.Lookuper) error {
	var err error
	ioregExpert, err = GetIORegTree("IOPlatformExpertDevice", l)
	if err != nil {
		return err
	}

	// Mmmm, return data
	return nil
}

func (o OSLookup) GetInstalledSoftware(l *lookups.Lookuper) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (o OSLookup) GetHostname(l *lookups.Lookuper) (interface{}, error) {
	out, err := l.Commander.Output("/usr/sbin/scutil", "--get", "LocalHostName")
	if err != nil {
		log.Warn().Msg("Error running 'scutil --get LocalHostName' to determine the hostname")
		return nil, err
	}
	trimmed := strings.Trim(string(out), "\n")
	return trimmed, nil
}

type SPSoftwareData struct {
	SPSoftwareDataType []struct {
		Name            string `json:"_name,omitempty"`
		BootMode        string `json:"boot_mode,omitempty"`
		BootVolume      string `json:"boot_volume,omitempty"`
		KernelVersion   string `json:"kernel_version,omitempty"`
		LocalHostName   string `json:"local_host_name,omitempty"`
		OsVersion       string `json:"os_version,omitempty"`
		SystemIntegrity string `json:"system_integrity,omitempty"`
		Uptime          string `json:"uptime,omitempty"`
		UserName        string `json:"user_name,omitempty"`
	}
}

type SPApplicationData struct {
	SPApplicationsDataType []struct {
		Name         string   `json:"_name,omitempty"`
		ArchKind     string   `json:"arch_kind,omitempty"`
		LastModified string   `json:"lastModified,omitempty"`
		ObtainedFrom string   `json:"obtained_from,omitempty"`
		Path         string   `json:"path,omitempty"`
		SignedBy     []string `json:"signed_by,omitempty"`
		Version      string   `json:"version,omitempty"`
	}
}

// This is like...OS data dude...
func GetPSoftwareData(l *lookups.Lookuper) (SPSoftwareData, error) {
	var s SPSoftwareData
	out, err := l.Commander.Output("/usr/sbin/system_profiler", "SPSoftwareDataType", "-json")
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(out, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// This is like...Application level data...my cool person
func GetPSApplicationData(l *lookups.Lookuper) (SPApplicationData, error) {
	var s SPApplicationData
	out, err := l.Commander.Output("/usr/sbin/system_profiler", "SPApplicationsDataType", "-json")
	if err != nil {
		log.Warn().Err(err).Msg("Error converting apps")
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(out, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func GetMemory(l *lookups.Lookuper) (int64, error) {
	memory, err := GetSysctl("hw.memsize", l)
	if err != nil {
		return 0, err
	}
	memoryMB := memory / 1024 / 1024

	return memoryMB, nil
}

func GetInstalledSoftware(l *lookups.Lookuper) ([][]string, error) {
	softwareTable := [][]string{}

	/*
		Homebrew packages reported here
	*/
	var brewp string
	if strings.HasPrefix(runtime.GOARCH, "arm") {
		brewp = "/opt/homebrew/bin/brew"
	} else {
		brewp = "/usr/local/bin/brew"
	}
	brewOut, err := l.Commander.Output(brewp, "list", "--versions")
	if err != nil {
		log.Warn().Msg("Homebrew package lookup failed")
	} else {
		trimmed := strings.Trim(string(brewOut), "\n")
		for _, line := range strings.Split(trimmed, "\n") {
			pieces := strings.Split(line, " ")
			name := pieces[0]
			version := pieces[1]

			softwareTable = append(softwareTable, []string{name, version})
		}
	}

	/*
		This is applications that MacOS knows about. The data is a little
		inconsistent as many packages don't list a version. Right now we are
		reporting the 'name', which may also be misleading. A more unique
		identifer might be 'path' for this...we should think about what the
		right way to report back is
	*/
	// Application Data
	data, err := GetPSApplicationData(l)
	if err != nil {
		return nil, err
	}
	if err != nil {
		log.Warn().Err(err).Msg("Could not get app date")
		return nil, err
	}
	for _, item := range data.SPApplicationsDataType {
		softwareTable = append(softwareTable, []string{item.Name, item.Version})
	}

	return softwareTable, nil
}

func GetSysctl(target string, l *lookups.Lookuper) (int64, error) {
	out, err := l.Commander.Output("/usr/sbin/sysctl", "-n", target)
	if err != nil {
		return 0, err
	}
	outClean := strings.TrimSuffix(string(out), "\n")

	v, err := strconv.ParseInt(outClean, 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("Error doing sysctl")
		return 0, err
	}
	return v, nil
}

func GetIORegTree(tree string, l *lookups.Lookuper) (map[string]string, error) {
	r := make(map[string]string)
	// out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", tree).Output()
	out, err := l.Commander.Output("/usr/sbin/ioreg", "-rd1", "-c", tree)
	if err != nil {
		return r, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		stripLine := strings.TrimSpace(line)
		if !strings.HasPrefix(stripLine, "\"") {
			continue
		}
		pieces := strings.Split(stripLine, " = ")
		// Strip off head and tail "s
		key := pieces[0]
		key = strings.ReplaceAll(key, "\"", "")

		// Strip < > from value
		value := pieces[1]
		value = strings.TrimLeft(value, "<")
		value = strings.TrimRight(value, ">")
		value = strings.ReplaceAll(value, "\"", "")
		r[key] = value
	}
	return r, nil
}

func GetDiskEncryptionStatus() (bool, error) {
	// Note fdsetup fails if the encryption is inactive
	_, err := exec.Command("/usr/bin/fdesetup", "isactive").Output()
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (o OSLookup) GetSerial(_ *lookups.Lookuper) (interface{}, error) {
	return ioregExpert["IOPlatformSerialNumber"], nil
}

func (o OSLookup) GetManufacturer(_ *lookups.Lookuper) (interface{}, error) {
	return ioregExpert["manufacturer"], nil
}

func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookuper) (interface{}, error) {
	encrypted, err := GetDiskEncryptionStatus()
	if err != nil {
		log.Warn().Err(err).Msg("Could not detect disk encryption state")
	}
	return encrypted, nil
}

func (o OSLookup) GetMemory(l *lookups.Lookuper) (interface{}, error) {
	memory, err := GetMemory(l)
	if err != nil {
		log.Warn().Msg("Could not detect memory")
	}
	return uint64(memory), nil
}

func extractCrowdstrikeAID(output []byte) (string, error) {
	// macOS prints this out with a bunch of junk...tryin to do this efficently
	// Looking up "agentID: <ActualID>\n"
	startMarker := "agentID: "
	aidStart := strings.Index(string(output), startMarker)
	aidPrefix := string(output)[aidStart+len(startMarker):]
	aidEnd := strings.Index(aidPrefix, "\n")
	aid := aidPrefix[:aidEnd]
	_, err := uuid.Parse(aid)
	if err != nil {
		return "", err
	}
	// Use the linux aid format, no uppercase/dashes
	aid = strings.ReplaceAll(strings.ToLower(aid), "-", "")
	return aid, nil
}

func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookuper) (interface{}, error) {
	ids := map[string]string{}
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	if user.Uid == "0" {
		aidOut, err := l.Commander.Output("/Applications/Falcon.app/Contents/Resources/falconctl", "stats", "agent_info")
		if err != nil {
			return nil, err
		}
		// macOS prints this out with a bunch of junk...tryin to do this efficently
		// Looking up "agentID: <ActualID>\n"
		aid, err := extractCrowdstrikeAID(aidOut)
		if err != nil {
			return nil, err
		}
		ids["crowdstrike_aid"] = aid
	}
	return ids, nil
}
