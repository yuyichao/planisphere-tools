package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gopkg.in/yaml.v2"
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

		overrides := viper.GetStringMap("overrides")

		l, err := helpers.NewLookuper(overrides)
		if err != nil {
			log.Fatal("Could not initialize Lookuper 😭☠️")
		}

		// Print payload
		out, _ := yaml.Marshal(l.Payload)

		fmt.Println(string(out))

		if !dryrun {
			err := l.Payload.Submit(planisphereKey)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Submitted report, thanks for keeping Duke Safe! ❤️")
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
	rootCmd.AddCommand(reportCmd)
}
