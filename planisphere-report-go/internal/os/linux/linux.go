/*
Package linux defines how to interact with the Linux os stuff
*/
package linux

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	report "gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

var (
	// ErrEmptyOutput is when there are no actual lines outputted
	ErrEmptyOutput = fmt.Errorf("empty output")
	// ErrMalformedPackageOutput is when the package output is bad
	ErrMalformedPackageOutput = fmt.Errorf("malformed Package Output")
)

// OSLookup handles the Linux OS lookups
type OSLookup struct{}

// ApplyPlatformDetections satisfies the OSLookuper interface
func (o OSLookup) ApplyPlatformDetections(_ *lookups.Lookup) error {
	return nil
}

// GetEnabledRepos returns a list of enabled repositories on a given system
func GetEnabledRepos(l *lookups.Lookup) ([]string, error) {
	if _, err := l.Commander.LookPath("apt-cache"); err == nil {
		return aptRepos(l)
	}
	if _, err := l.Commander.LookPath("yum"); err == nil {
		return yumRepos(l)
	}
	return nil, errors.New("no apt-cache or yum found to look up enabled repositories")
}

// GetInstalledSoftware returns the installed software
func GetInstalledSoftware(l *lookups.Lookup) ([][]string, error) {
	softwareTable := [][]string{}
	softwareQueries := [][]string{
		{"rpm", "-qa", "--qf", "%{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n"},
		{"guix-installed"},
		{"dpkg-query", "-W"},
		{"pacman", "-Q"},
	}

	for _, softwareQuery := range softwareQueries {
		cmd := softwareQuery[0]
		args := softwareQuery[1:]
		binPath, err := l.Commander.LookPath(cmd)
		if err == nil {
			slog.Debug("running software query", "query", softwareQuery)
			binOut, err := l.Commander.Output(binPath, args...)
			if err != nil {
				slog.Warn("could not do a listing even though the rpm command exists", "command", cmd)
			}
			binSoftware, err := ParsePackageOutput(binOut)
			if err != nil {
				slog.Warn("could not parse the command output", "command", cmd)
			} else {
				softwareTable = append(softwareTable, binSoftware...)
			}
		} else {
			slog.Debug("command installed", "cmd", cmd)
		}
	}

	return softwareTable, nil
}

// GetExtendedOSSupport returns the vendor providing Extended OS Support for an operating system
func (o OSLookup) GetExtendedOSSupport(l *lookups.Lookup) (interface{}, error) {
	repos, err := GetEnabledRepos(l)
	if err != nil {
		slog.Debug("could not find any repos to do an extended os support lookup on")
		return "", nil
	}

	for _, repo := range repos {
		for sp, regexes := range map[string][]regexp.Regexp{
			"Red Hat ELS": {
				*regexp.MustCompile(`\/\/yum.oit.duke.edu\/patchmonkey\/el\d+\/rhel-\d+-server-els-rpms`),
			},
			"TuxCare": {
				*regexp.MustCompile(`\/\/yum.oit.duke.edu\/patchmonkey\/el\d+\/centos-\d+-els\/`),
				*regexp.MustCompile(`\/\/repo.tuxcare.com`),
			},
			"Ubuntu Pro": {
				*regexp.MustCompile(`\/\/apt.oit.duke.edu\/dists\/\S+-infra-(updates|security).*`),
				*regexp.MustCompile(`\/\/esm.ubuntu.com\/`),
			},
		} {
			for _, re := range regexes {
				if re.MatchString(repo) {
					slog.Debug("found repo matching a support regex", "repo", repo, "support-provider", sp, "regex", re)
					return sp, nil
				}
			}
		}
	}
	return "", nil
}

// GetSerial satisfies the OSLookuper interface
func (o OSLookup) GetSerial(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Slurp("/sys/class/dmi/id/product_serial")
	if err != nil {
		return nil, err
	}
	serial := strings.TrimSpace(string(out))
	return serial, nil
}

// GetManufacturer satisfies the OSLookuper interface
func (o OSLookup) GetManufacturer(l *lookups.Lookup) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			return "Raspberry Pi", nil
		}
	}

	// bios_vendor isn't truly accurate on the linux hosts, moving to
	// sys_vendor which seems to be more inline with the other EPM tools
	out, err := l.Commander.Slurp("/sys/class/dmi/id/sys_vendor")
	if err != nil {
		return nil, err
	}
	vendor := strings.TrimSuffix(string(out), "\n")
	return vendor, nil
}

// GetModel satisfies the OSLookuper interface
func (o OSLookup) GetModel(l *lookups.Lookup) (interface{}, error) {
	// Is it a Raspberry Pi?
	l.WaitForChecked("mac_addresses")
	for _, mac := range l.Payload.Data.MacAddresses {
		if strings.HasPrefix(mac, "b8:27:eb") {
			cpuDat, err := l.Commander.Slurp("/proc/cpuinfo")
			if err != nil {
				slog.Warn("could not get Raspberry Pi CPU info", "error", err)
				continue
			}
			trimmed := strings.Trim(string(cpuDat), "\n")
			for _, line := range strings.Split(trimmed, "\n") {
				pieces := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(pieces[0])
				value := strings.TrimSpace(pieces[1])
				if key == "Revision" {
					if _, ok := report.RaspberryPiModels[value]; ok {
						return report.RaspberryPiModels[value], nil
					}
				}
			}
		}
	}
	out, err := l.Commander.Slurp("/sys/class/dmi/id/product_name")
	if err != nil {
		return nil, err
	}
	productName := strings.TrimSuffix(string(out), "\n")
	return productName, nil
}

// GetDiskEncrypted satisfies the OSLookuper interface
func (o OSLookup) GetDiskEncrypted(_ *lookups.Lookup) (interface{}, error) {
	return false, errors.New("DiskEncrypted Not yet implemented")
}

// GetMemory satisfies the OSLookuper interface
func (o OSLookup) GetMemory(l *lookups.Lookup) (interface{}, error) {
	out, err := l.Commander.Slurp("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	reMemTotal := regexp.MustCompile(`^MemTotal:\s+(\d+)\s+.*$`)
	var memory uint64
	for s.Scan() {
		if m := reMemTotal.FindStringSubmatch(s.Text()); m != nil {
			memory, err = strconv.ParseUint(m[1], 10, 64)
			if err != nil {
				return nil, err
			}
		}
	}
	// Switch to MB here
	return memory / 1024, err
}

// GetOSFamily satisfies the OSLookuper interface
func (o OSLookup) GetOSFamily(_ *lookups.Lookup) (interface{}, error) {
	return "Linux", nil
}

// GetDeviceType satisfies the OSLookuper interface
func (o OSLookup) GetDeviceType(l *lookups.Lookup) (interface{}, error) {
	l.WaitForChecked("serial")
	if strings.Contains(l.Payload.Data.Serial, "VMware") {
		return "vm", nil
	}
	ct, err := l.Commander.Slurp("/sys/class/dmi/id/chassis_type")
	if err != nil {
		return nil, err
	}
	ctt := strings.TrimSpace(string(ct))
	cti, err := strconv.Atoi(ctt)
	if err != nil {
		return nil, err
	}
	// Is the chassis type integer in the deviceTypeTable?
	if _, ok := report.ChassisType[cti]; ok {
		return report.ChassisType[cti], nil
	}

	return "", nil
}

// GetOSFullName satisfies the OSLookuper interface
func (o OSLookup) GetOSFullName(l *lookups.Lookup) (interface{}, error) {
	osb, err := l.Commander.Slurp("/etc/os-release")
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(bytes.NewReader(osb))
	var fullName string

	rePrettyName := regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil {
			fullName = strings.Trim(m[1], `"`)
		}
	}
	return fullName, nil
}

func aptRepos(l *lookups.Lookup) ([]string, error) {
	out, err := l.Commander.Output("/usr/bin/apt-cache", "policy")
	if err != nil {
		return nil, errors.New("could not run apt-cache")
	}
	ret := []string{}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, "http") {
			pieces := strings.Split(txt, " ")
			ret = append(ret, fmt.Sprintf("%v/dists/%v", pieces[2], pieces[3]))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Strings(ret)
	return ret, nil
}

func yumRepos(l *lookups.Lookup) ([]string, error) {
	out, err := l.Commander.Output("/usr/bin/yum", "repolist", "-v", "enabled")
	if err != nil {
		return nil, errors.New("could not run yum")
	}
	r := regexp.MustCompile(`^Repo-(baseurl|mirrors)\s+:\s(\S+).*$`)

	ret := []string{}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		txt := scanner.Text()
		match := r.FindStringSubmatch(txt)
		if len(match) > 0 {
			ret = append(ret, match[2])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Strings(ret)
	return ret, nil
}

// GetInstalledSoftware satisfies the OSLookuper interface
func (o OSLookup) GetInstalledSoftware(l *lookups.Lookup) (interface{}, error) {
	apps, err := GetInstalledSoftware(l)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// GetHostname satisfies the OSLookuper interface
func (o OSLookup) GetHostname(_ *lookups.Lookup) (interface{}, error) {
	// Hostname Field
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return hostname, nil
}

// GetExternalOSIdentifiers satisfies the OSLookuper interface
func (o OSLookup) GetExternalOSIdentifiers(l *lookups.Lookup) (interface{}, error) {
	ids := map[string]string{}
	aidOut, err := l.Commander.Output("/opt/CrowdStrike/falconctl", "-g", "--aid")
	if err != nil {
		return nil, err
	}
	ids["crowdstrike_aid"] = strings.TrimSuffix(strings.TrimPrefix(string(aidOut), "aid=\""), "\".\n")
	return ids, nil
}

// MakeNamePro inserts 'Pro' after the first part of the string
// so "Ubuntu 18.04" becomes "Ubuntu Pro 18.04"
func MakeNamePro(s string) string {
	pieces := strings.Split(s, " ")
	return fmt.Sprintf("%s Pro %s", pieces[0], strings.Join(pieces[1:], " "))
}
