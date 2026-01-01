// Package cmd contains the CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/akz4ol/gatewayops/cli/internal/api"
)

var (
	cfgFile string
	apiKey  string
	baseURL string
	client  *api.Client
)

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "gwo",
	Short: "GatewayOps CLI - Manage MCP Gateway operations",
	Long: `GatewayOps CLI provides commands to interact with the GatewayOps MCP Gateway.

You can manage API keys, call MCP tools, view traces, and monitor costs.

Get started by setting your API key:
  gwo auth login

Or set the GATEWAYOPS_API_KEY environment variable.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip auth check for auth commands
		if cmd.Name() == "login" || cmd.Name() == "logout" || cmd.Name() == "version" {
			return nil
		}

		// Get API key from flag, env, or config
		key := apiKey
		if key == "" {
			key = viper.GetString("api_key")
		}
		if key == "" {
			key = os.Getenv("GATEWAYOPS_API_KEY")
		}

		if key == "" {
			return fmt.Errorf("API key not set. Run 'gwo auth login' or set GATEWAYOPS_API_KEY")
		}

		// Get base URL
		url := baseURL
		if url == "" {
			url = viper.GetString("base_url")
		}
		if url == "" {
			url = os.Getenv("GATEWAYOPS_BASE_URL")
		}

		client = api.NewClient(key, url)
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gwo.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "GatewayOps API key")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "API base URL")

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(tracesCmd)
	rootCmd.AddCommand(costsCmd)
	rootCmd.AddCommand(versionCmd)
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
		// Config file found
	}
}

// versionCmd shows the CLI version.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gwo version 0.1.0")
	},
}
