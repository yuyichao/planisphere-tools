/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/apex/log"
	"gitlab.oit.duke.edu/devil-ops/planisphere-sdk/planisphere"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/gpg"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// diagnosticsCmd represents the diagnostics command
var diagnosticsCmd = &cobra.Command{
	Use:     "diagnostics",
	Aliases: []string{"d", "diag", "diagnostic"},
	Short:   "Run commands locally to generate a diagnostic report you can send to the maintainers when things go wonky",
	Long: `This command will run some local commands and capture the output. This should be
used when your system is not generating accurate information, or if things are
missing or erroring.

Default behavior is to encrypt this output to the ssi-systems team. If you would
like to save it as plaintext, use the --plaintext flag. Feel free to re-encrypt
it with whatever public keys you choose as well.`,

	Run: func(cmd *cobra.Command, args []string) {
		plaintext, err := cmd.Flags().GetBool("plaintext")
		cobra.CheckErr(err)
		stdout, err := cmd.Flags().GetBool("stdout")
		cobra.CheckErr(err)
		type cmdDiagnostic struct {
			Command string `yaml:"command"`
			StdOut  string `yaml:"stdout"`
			StdErr  string `yaml:"stderr"`
		}
		type diagnostic struct {
			Hostname  string                        `yaml:"hostname"`
			Version   string                        `yaml:"version"`
			User      string                        `yaml:"user"`
			TimeStamp time.Time                     `yaml:"time"`
			Instance  string                        `yaml:"instance"`
			KeyHash   string                        `yaml:"key_hash"`
			Overrides map[string]interface{}        `yaml:"overrides"`
			Payload   planisphere.SelfReportPayload `yaml:"payload"`
			Commands  []cmdDiagnostic               `yaml:"commands"`
		}
		cmdMap := map[string][][]string{
			"darwin": {
				{"/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice"},
				{"/usr/bin/fdesetup", "isactive"},
				{"/usr/sbin/scutil", "--get", "LocalHostName"},
				{"/usr/sbin/system_profiler", "SPSoftwareDataType", "-json"},
				{"/usr/sbin/system_profiler", "SPApplicationsDataType", "-json"},
				{"/usr/local/bin/brew", "list", "--versions"},
				{"/usr/sbin/sysctl", "-a"},
				{"/Applications/Falcon.app/Contents/Resources/falconctl", "stats", "agent_info"},
				{"/usr/local/bin/brew", "config"},
				{"/usr/local/bin/brew", "services"},
				{"/bin/cat", "/usr/local/var/log/planisphere-report.err"},
			},
			"linux": {
				{"/opt/CrowdStrike/falconctl", "-g", "--aid"},
				{"dmidecode", "-s", "bios-vendor"},
				{"dmidecode", "-s", "system-product-name"},
				{"dmidecode", "-s", "system-serial-number"},
				{"dmidecode", "-s", "chassis-type"},
				{"/usr/bin/rpm", "-qa", "--qf", "%{NAME} %|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}"},
				{"/usr/bin/dpkg-query", "-W"},
				{"/usr/bin/guix-installed"},
				{"/usr/bin/pacman", "-Q"},
			},
		}

		cmds := cmdMap[runtime.GOOS]
		if cmds == nil {
			log.WithField("os", runtime.GOOS).Fatal("No commands available for this OS")
		}

		hostname, _ := os.Hostname()
		now := time.Now()
		d := diagnostic{
			Hostname:  hostname,
			Version:   version,
			Instance:  planisphereURL,
			TimeStamp: now,
			KeyHash:   helpers.HashString(planisphereKey),
			Overrides: viper.GetStringMap("overrides"),
		}
		u, err := user.Current()
		if err != nil {
			log.WithError(err).Warn("Could not get username")
		} else {
			d.User = u.Username
		}
		c := make(chan cmdDiagnostic)
		for _, cmd := range cmds {
			go func(cmd []string) {
				oCmd := exec.Command(cmd[0], cmd[1:]...)
				var outbuf, errbuf strings.Builder
				oCmd.Stdout = &outbuf
				oCmd.Stderr = &errbuf

				err := oCmd.Run()
				if err != nil {
					log.WithError(err).WithField("cmd", cmd).Debug("Error running command")
				}
				oCmd.Wait()

				cd := cmdDiagnostic{
					Command: strings.Join(cmd, " "),
					StdOut:  outbuf.String(),
					StdErr:  errbuf.String(),
				}
				c <- cd
			}(cmd)
		}
		for i := 0; i < len(cmds); i++ {
			d.Commands = append(d.Commands, <-c)
		}

		// Try to get a real report as well, and toss it in for good measure
		lc := &lookups.LookuperConfig{
			Overrides: viper.GetStringMap("overrides"),
		}
		lu, err := helpers.NewLookuper(lc)
		if err == nil {
			d.Payload = lu.Payload
		} else {
			log.WithError(err).Warn("Issues getting lookups")
		}

		// Marshal the diagnostic
		b, err := yaml.Marshal(d)
		cobra.CheckErr(err)

		// Write out to temp file in gzip
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		zw.Name = "planisphere-report-diagnostic"
		zw.Comment = "Output of planisphere-report diagnostic"

		// Do we want to encrypt?
		if !plaintext {
			els, err := gpg.CollectGPGPubKeys("")
			cobra.CheckErr(err)
			b, err = gpg.Encrypt(b, els)
			cobra.CheckErr(err)
		}

		// Write data out
		zw.Write(b)
		zw.Close()

		// fmt.Println(string(encrypted))

		var fExt string
		if plaintext {
			fExt = "*.yaml.gz"
		} else {
			fExt = "*.yaml.gpg.gz"
		}
		diagFile, err := os.CreateTemp("", fExt)
		cobra.CheckErr(err)
		err = ioutil.WriteFile(diagFile.Name(), buf.Bytes(), os.ModePerm)
		cobra.CheckErr(err)
		fmt.Printf("Wrote diagnostics to: %v\n", diagFile.Name())
		fmt.Println("Please describe the issue you are running in to, and attach this diagnostic file to a new 'Issue' report at the URL below")
		fmt.Println("https://duke.is/z9nxa")
		if stdout {
			fmt.Println(string(b))
		}
	},
}

func init() {
	rootCmd.AddCommand(diagnosticsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// diagnosticsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	diagnosticsCmd.PersistentFlags().BoolP("plaintext", "p", false, "Use plaintext, don't encrypt")
	diagnosticsCmd.PersistentFlags().BoolP("stdout", "s", false, "Send output to stdout in addition to the tmpfile")
}
