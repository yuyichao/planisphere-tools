package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
	"gopkg.in/yaml.v2"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report back to Planisphere",
	Long:  `Look up local information and send it up to Planisphere self report`,
	Run: func(cmd *cobra.Command, args []string) {
		// Note: We are printing the time to stdout so that we can
		// collect this in the non-error log, which is stdout on macOS
		logger.Info("starting report collection")
		dryrun, err := cmd.Flags().GetBool("dryrun")
		cobra.CheckErr(err)

		hideSummary, err := cmd.Flags().GetBool("hide-summary")
		cobra.CheckErr(err)

		interval, err := cmd.Flags().GetDuration("interval")
		cobra.CheckErr(err)

		// sbomTarget, err := cmd.Flags().GetStringArray("sbom-target")
		// cobra.CheckErr(err)

		for {
			err := os.Setenv("PLANISPHEREREPORT_URL", planisphereURL)
			if err != nil {
				logFatal("error setting url", err)
			}

			c := &lookups.LookuperConfig{
				Overrides: viper.GetStringMap("overrides"),
			}
			l, err := helpers.NewLookuper(c)
			if err != nil {
				logFatal("Could not initialize Lookuper 😭☠️", err)
			}

			// Add version to extra data
			l.Payload.ExtraData["selfreport_version"] = version

			// Print payload when in verbose mode
			if Verbose {
				out, _ := yaml.Marshal(l.Payload)
				fmt.Println(string(out))
			} else if !hideSummary {
				// If asked, print a little summary out
				summaryText := strings.Builder{}
				// In normal mode, just show a summary
				e := reflect.ValueOf(&l.Payload.Data).Elem()
				for i := 0; i < e.NumField(); i++ {
					varName := e.Type().Field(i).Name
					varValue := e.Field(i).Interface()
					switch name := varName; name {
					case "InstalledSoftware":
						summaryText.WriteString(fmt.Sprintf("%v: %v items\n", varName, len(varValue.([][]string))))
					default:
						summaryText.WriteString(fmt.Sprintf("%v: %v\n", varName, varValue))

					}
				}
				summaryText.WriteString(fmt.Sprintf("ExtraData: %+v\n", l.Payload.ExtraData))
				fmt.Print(summaryText.String())
				fmt.Printf("%v Completed report collection\n", time.Now())
			}

			if !dryrun {
				err := l.Payload.Submit(planisphereKey)
				if err != nil {
					logFatal("error submitting payload", err)
				}
				logger.Info("Submitted report, thanks for keeping Duke Safe! ❤️", "duration", fmt.Sprint(time.Since(startedAt)))
			}
			if interval.Seconds() == 0 {
				return
			}
			logger.Info("sleeping until next run", "interval", fmt.Sprint(interval))
			time.Sleep(interval)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	reportCmd.Flags().BoolP("dryrun", "d", false, "Do a dry run, don't actually submit to planisphere")
	reportCmd.Flags().Bool("hide-summary", false, "Don't print out a summary of the report")
	reportCmd.Flags().DurationP("interval", "i", 0*time.Second, "Instead of running once and exiting, run continually while sleeping at the given interval. Must be compatible with https://pkg.go.dev/time#ParseDuration")
	reportCmd.Flags().StringArray("sbom-target", []string{}, "Generate an SBOM for the given directory to include in the report")
	rootCmd.AddCommand(reportCmd)
}
