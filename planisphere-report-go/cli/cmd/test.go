package cmd

import (
	"fmt"
	"os"

	"github.com/r3labs/diff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/internal/lookups"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test self-report",
	Long: `Use the /self_report/test endpoint to see how the data you are submitting will
be interpreted by Planisphere. Note that this does not actually submit your self report. Use
the 'report' command for that`,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.Setenv("PLANISPHEREREPORT_URL", planisphereURL+"/test")
		if err != nil {
			log.Fatal("Error setting url: ", err)
		}
		dryrun, _ := cmd.Flags().GetBool("dryrun")
		log.Debug("Dryrun is set to: ", dryrun)

		overrides := viper.GetStringMap("overrides")

		c := &lookups.LookuperConfig{
			Overrides: overrides,
		}
		l, err := helpers.NewLookuper(c)
		if err != nil {
			log.Warning(err)
			log.Fatal("Could not initialize Lookuper 😭☠️")
		}

		// Add version to extra data
		l.Payload.ExtraData["selfreport_version"] = version
		payloadResp, err := l.Payload.SubmitTest(planisphereKey)
		if err != nil {
			log.Fatal(err)
		}
		if payloadResp.Status != "success" {
			log.Fatalf("Could not send this data in as test data, got back: %v", payloadResp)
		}
		changelog, _ := diff.Diff(l.Payload, payloadResp.ProcessedRecord)
		if len(changelog) > 0 {
			fmt.Println("After being interpreted by Planisphere, the following changes will be made to the data you are going to submit:")
			for _, change := range changelog {
				fmt.Printf("%+v\n", change)
			}
		} else {
			log.Println("Wooooow, no changes!")
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
