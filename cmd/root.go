package cmd

import (
	"fmt"
	"informant/internal/config"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "1.4.1" // Matching original version
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "informant",
	Short: "An Arch Linux News reader and pacman hook",
	Long: `informant is an Arch Linux News reader designed to be used as a pacman hook.
It can interrupt pacman transactions to ensure you have read the news first.

informant provides commands to check, list, and read news items, plus an
interactive TUI mode for browsing news.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.informantrc.json)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("no-confirm", false, "skip confirmation prompts for storage fallback")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no-confirm", rootCmd.PersistentFlags().Lookup("no-confirm"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory and standard locations
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			return
		}

		// Search config in multiple locations as per original informant
		viper.AddConfigPath(home)
		viper.AddConfigPath("$XDG_CONFIG_HOME")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath(".")
		viper.SetConfigType("json")
		viper.SetConfigName(".informantrc")
		viper.SetConfigName("informantrc")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	} else {
		// Initialize default config if no config file found
		config.SetDefaults()
	}
}
