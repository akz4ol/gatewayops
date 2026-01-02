package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiKey  string
	baseURL string
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "gwo",
	Short: "GatewayOps CLI - Manage your MCP Gateway",
	Long: `GatewayOps CLI provides a command-line interface to interact with
the GatewayOps MCP Gateway platform. Manage API keys, view traces,
monitor costs, and interact with MCP servers.

Configure with environment variables:
  GATEWAYOPS_API_KEY  - Your API key
  GATEWAYOPS_BASE_URL - API base URL (default: https://api.gatewayops.com)`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gwo.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "GatewayOps API key")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "https://api.gatewayops.com", "API base URL")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (table, json)")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gwo")
	}

	viper.SetEnvPrefix("GATEWAYOPS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return viper.GetString("api_key")
}

func getBaseURL() string {
	if baseURL != "" {
		return baseURL
	}
	url := viper.GetString("base_url")
	if url == "" {
		return "https://api.gatewayops.com"
	}
	return url
}
