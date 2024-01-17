package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/user"

	"github.com/spf13/cobra"
)

// installCronCmd represents the installCron command
var installCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Install cron runner",
	Run: func(cmd *cobra.Command, args []string) {
		cronFile, _ := cmd.Flags().GetString("cron-file")
		force, _ := cmd.Flags().GetBool("force")
		user, _ := cmd.Flags().GetString("user")
		if exists(cronFile) && !force {
			logger.Error("Cronfile already exists. Use -f/--force to overwrite it", "file", cronFile)
			os.Exit(2)
		}
		// Figure out out path
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}

		content := fmt.Sprintf("%v %v * * * %v %v report\n", mustRand(59), mustRand(23), user, ex)
		logger.Info("creating cron:")
		fmt.Println(content)

		cobra.CheckErr(os.WriteFile(cronFile, []byte(content), 0o600))
		logger.Info("successfully installed cron! If it's not to your liking, feel free to edit.")
	},
}

func mustRand(i int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(i))
	if err != nil {
		panic(err)
	}
	return n.Int64()
}

func init() {
	installCmd.AddCommand(installCronCmd)

	// Here you will define your flags and configuration settings.
	me, _ := user.Current()

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	installCronCmd.PersistentFlags().StringP("cron-file", "c", "/etc/cron.d/planisphere-report", "A help for foo")
	installCronCmd.Flags().BoolP("force", "f", false, "Force override of existing file")
	installCronCmd.Flags().String("user", me.Username, "User to run the cron as")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCronCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
