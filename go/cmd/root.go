package cmd

import (
	"github.com/furlongm/patchman-client/go/lib/report"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:     "patchman-client",
		Short:   "A Go client for Patchman",
		Version: report.Version,
		Long: `Patchman is a Django-based Linux patch management system.
This Go client collects and submits host details to a Patchman server.`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/patchman/patchman-client.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug output")
	rootCmd.PersistentFlags().StringP("server", "s", "", "patchman server URL")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "API key for authentication")
	rootCmd.PersistentFlags().Bool("no-verify-ssl", false, "disable SSL certificate verification")
	rootCmd.PersistentFlags().String("ca-file", "", "path to custom CA certificate file")
	rootCmd.PersistentFlags().StringP("tags", "t", "", "tags to assign to this host")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("no_verify_ssl", rootCmd.PersistentFlags().Lookup("no-verify-ssl"))
	viper.BindPFlag("ca_file", rootCmd.PersistentFlags().Lookup("ca-file"))
	viper.BindPFlag("tags", rootCmd.PersistentFlags().Lookup("tags"))

	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(submitCmd)
}

// MaskAPIKey masks an API key for safe display in logs
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/patchman")
		viper.SetConfigType("yaml")
		viper.SetConfigName("patchman-client")
	}

	viper.AutomaticEnv()

	if viper.GetBool("verbose") {
		jww.SetStdoutThreshold(jww.LevelInfo)
	}

	if viper.GetBool("debug") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}

	if err := viper.ReadInConfig(); err == nil {
		jww.INFO.Println("Using config file:", viper.ConfigFileUsed())
	}
}
