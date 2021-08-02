package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var planisphereKey string
var planisphereURL string

// Verbose Logging
var Verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "planisphere-report",
	Short: "Self report tool for Planisphere",
	Long:  `Self report tool for Planisphere. Currently only supported on Macos`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//planisphereKey := viper.GetString("key")
		// Are we talky?
		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		// Key is gonna be required
		planisphereKey = viper.GetString("key")
		if planisphereKey == "" {
			log.Fatal("Must set your Planisphere Key in the config file or env. See README.md for details")
		}
		log.Debug("Using Key: ", planisphereKey)

		// Set URL if needed
		planisphereURL = viper.GetString("url")
		if planisphereURL == "" {
			planisphereURL = "https://planisphere.oit.duke.edu/self_report"
		}
		log.Debug("Using URL: ", planisphereURL)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.planisphere-report.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".planisphere-report" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".planisphere-report")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("planispherereport")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}
}
