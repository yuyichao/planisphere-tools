package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"gitlab.oit.duke.edu/devil-ops/planisphere-tools/planisphere-report-go/helpers"
)

// installCronCmd represents the installCron command
var installCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Install cron runner",
	Run: func(cmd *cobra.Command, args []string) {
		cronFile, _ := cmd.Flags().GetString("cron-file")
		force, _ := cmd.Flags().GetBool("force")
		if helpers.Exists(cronFile) && !force {
			log.Fatalf("Cronfile %v already exists. Use -f/--force to overwrite it", cronFile)
		}
		// Figure out out path
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}

		// Semi random seeding
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		hour := r1.Intn(23)
		minute := r1.Intn(59)
		content := fmt.Sprintf("%v %v * * * %v report\n", minute, hour, ex)
		log.Println("Creating cron:")
		fmt.Println(content)

		err = os.WriteFile(cronFile, []byte(content), 0o644)

		cobra.CheckErr(err)
		log.Println("Successfully installed cron! If it's not to your liking, feel free to edit.")
	},
}

func init() {
	installCmd.AddCommand(installCronCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	installCronCmd.PersistentFlags().StringP("cron-file", "c", "/etc/cron.d/planisphere-report", "A help for foo")
	installCronCmd.Flags().BoolP("force", "f", false, "Force override of existing file")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCronCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
