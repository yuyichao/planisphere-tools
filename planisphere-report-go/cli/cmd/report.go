package cmd

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
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
		dryrun, err := cmd.Flags().GetBool("dryrun")
		cobra.CheckErr(err)
		log.Debug().Bool("dryrun", dryrun).Msg("Dryrun mode")

		interval, err := cmd.Flags().GetDuration("interval")
		cobra.CheckErr(err)

		for {
			err := os.Setenv("PLANISPHEREREPORT_URL", planisphereURL)
			if err != nil {
				log.Fatal().Err(err).Msg("Error setting url")
			}

			c := &lookups.LookuperConfig{
				Overrides: viper.GetStringMap("overrides"),
			}
			l, err := helpers.NewLookuper(c)
			if err != nil {
				log.Fatal().Err(err).Msg("Could not initialize Lookuper 😭☠️")
			}

			// Add version to extra data
			l.Payload.ExtraData["selfreport_version"] = version

			// Print payload when in verbose mode
			if Verbose {
				out, _ := yaml.Marshal(l.Payload)
				fmt.Println(string(out))

			} else {
				// In normal mode, just show a summary
				e := reflect.ValueOf(&l.Payload.Data).Elem()
				for i := 0; i < e.NumField(); i++ {
					varName := e.Type().Field(i).Name
					varValue := e.Field(i).Interface()
					switch name := varName; name {
					case "InstalledSoftware":
						fmt.Printf("%v: %v items\n", varName, len(varValue.([][]string)))
					default:
						fmt.Printf("%v: %v\n", varName, varValue)

					}
				}
				fmt.Printf("ExtraData: %+v\n", l.Payload.ExtraData)

			}

			if !dryrun {
				err := l.Payload.Submit(planisphereKey)
				if err != nil {
					log.Fatal().Err(err).Msg("Error submitting payload")
				}
				fmt.Println("Submitted report, thanks for keeping Duke Safe! ❤️")
			}
			if interval.Seconds() == 0 {
				return
			}
			log.Info().Str("interval", fmt.Sprint(interval)).Msg("Sleeping until next run")
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
	reportCmd.Flags().DurationP("interval", "i", 0*time.Second, "Instead of running once and exiting, run continually while sleeping at the given interval. Must be compatible with https://pkg.go.dev/time#ParseDuration")
	rootCmd.AddCommand(reportCmd)
}
