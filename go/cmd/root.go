package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

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
	rootCmd.PersistentFlags().BoolP("local-updates", "u", false, "find updates locally (e.g. for RHEL)")
	rootCmd.PersistentFlags().BoolP("no-repo-check", "n", false, "skip repository collection")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("no_verify_ssl", rootCmd.PersistentFlags().Lookup("no-verify-ssl"))
	viper.BindPFlag("ca_file", rootCmd.PersistentFlags().Lookup("ca-file"))
	viper.BindPFlag("tags", rootCmd.PersistentFlags().Lookup("tags"))
	viper.BindPFlag("local_updates", rootCmd.PersistentFlags().Lookup("local-updates"))
	viper.BindPFlag("no_repo_check", rootCmd.PersistentFlags().Lookup("no-repo-check"))

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

	if err := viper.ReadInConfig(); err != nil {
		// YAML not found — try shell format .conf as fallback
		confPath := "/etc/patchman/patchman-client.conf"
		loadShellConfig(confPath)
		// Load conf.d drop-in files
		loadConfDir(filepath.Dir(confPath))
	} else {
		jww.INFO.Println("Using config file:", viper.ConfigFileUsed())
		// Load conf.d drop-in files alongside YAML config
		loadConfDir(filepath.Dir(viper.ConfigFileUsed()))
	}
}

// loadConfDir loads all files from <dir>/conf.d/ as shell-format configs.
func loadConfDir(dir string) {
	confDir := filepath.Join(dir, "conf.d")
	entries, err := os.ReadDir(confDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			loadShellConfig(filepath.Join(confDir, entry.Name()))
		}
	}
}

// loadShellConfig reads a shell-format config file (KEY=value lines)
// for compatibility with the bash/python patchman-client.conf.
func loadShellConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	jww.INFO.Println("Using config file:", path)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, "\"'")

		// Map shell config keys to viper keys
		switch key {
		case "server":
			viper.SetDefault("server", value)
		case "api_key":
			viper.SetDefault("api_key", value)
		case "tags":
			viper.SetDefault("tags", value)
		case "verbose":
			viper.SetDefault("verbose", parseBool(value))
		case "debug":
			viper.SetDefault("debug", parseBool(value))
		case "local_updates":
			viper.SetDefault("local_updates", parseBool(value))
		case "repo_check":
			// repo_check=false in config → no_repo_check=true
			viper.SetDefault("no_repo_check", !parseBool(value))
		}
	}
}

// parseBool interprets shell-style boolean values.
func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes"
}
