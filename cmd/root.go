package cmd

import (
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "patchman-client",
		Short: "A go client for patchman",
		Long: `Patchman is a django-based linux patch management system.
This go client collects and submits a hosts details to a Patchman server.`,
	}
	verbose bool
	debug   bool
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/patchman/patchman-client.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "debug output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	// local-updates, tags, server, hostname, config, repo-check
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(submitCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir := "/etc/patchman/patchman-client.yaml"
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("patchman-client")
	}

	viper.AutomaticEnv()

	if verbose {
		jww.SetStdoutThreshold(jww.LevelInfo)
	}

	if debug {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}

	if err := viper.ReadInConfig(); err == nil {
		jww.INFO.Println("Using config file:", viper.ConfigFileUsed())
	}
}
