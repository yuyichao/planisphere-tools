package cmd

import (
	"C"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
)
import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report back to Planisphere",
	Long:  `Look up local information and send it up to Planisphere self report`,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.Setenv("PLANISPHEREREPORT_URL", planisphereURL)
		if err != nil {
			log.Fatal("Error setting url: ", err)
		}
		dryrun, _ := cmd.Flags().GetBool("dryrun")
		log.Debug("Dryrun is set to: ", dryrun)
		payload := &planisphere.SelfReportPayload{}

		// Hostname Field
		hostname, err := os.Hostname()
		if err != nil {
			log.Warning("Could not detect hostname: ", err)
		}
		payload.Data.Hostname = hostname

		// Mac Addresses Field
		macs, err := getMacAddr()
		if err != nil {
			log.Warning("Could not detect mac_addresses: ", err)
		}
		payload.Data.MacAddresses = macs

		// Memory here
		memory, err := getSysctl("hw.memsize")
		if err != nil {
			log.Warning("Could not detect memory size: ", err)
		}
		memoryMB := memory / 1024
		payload.Data.MemoryMB = memoryMB

		// Detect Serial Number
		serial, err := GetIORegValue("IOPlatformExpertDevice", "IOPlatformUUID")
		if err != nil {
			log.Warning("Could not detect serial: ", err)
		}
		payload.Data.Serial = serial

		// Set last active
		payload.LastActive = time.Now()

		// Print payload
		log.Printf("%+v", payload)
		if !dryrun {
			err := payload.Submit(planisphereKey)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Submitted report, thanks for keeping Duke Safe! ❤️")
		}

	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	reportCmd.Flags().BoolP("dryrun", "d", false, "Do a dry run, don't actually submit to planisphere")
}

func getSysctl(target string) (int64, error) {
	out, err := exec.Command("/usr/sbin/sysctl", "-n", target).Output()
	if err != nil {
		return 0, err
	}
	outClean := strings.TrimSuffix(string(out), "\n")

	v, err := strconv.ParseInt(outClean, 10, 64)
	if err != nil {
		log.Warning("Error doing sysctl: ", err)
		return 0, err
	}
	return v, nil
}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}

func GetIORegValue(tree, item string) (string, error) {
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", tree).Output()
	if err != nil {
		return "", err
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
		if key == item {
			return value, nil
		}
	}
	return "", nil

}
