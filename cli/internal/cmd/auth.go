package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Manage authentication for the GatewayOps CLI.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		
		if key == "" {
			fmt.Print("Enter your GatewayOps API key: ")
			fmt.Scanln(&key)
		}

		if key == "" {
			return fmt.Errorf("API key is required")
		}

		// Save to config file
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configPath := filepath.Join(home, ".gwo.yaml")
		config := map[string]string{
			"api_key": key,
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Println("API key saved to ~/.gwo.yaml")
		fmt.Println("You can now use gwo commands without --api-key flag")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		key := getAPIKey()
		if key == "" {
			fmt.Println("Not authenticated. Run 'gwo auth login' to configure.")
			return nil
		}

		// Mask the key
		masked := key[:8] + "..." + key[len(key)-4:]
		fmt.Printf("Authenticated with key: %s\n", masked)
		fmt.Printf("Base URL: %s\n", getBaseURL())
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configPath := filepath.Join(home, ".gwo.yaml")
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config: %w", err)
		}

		fmt.Println("Logged out. Config file removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	authLoginCmd.Flags().StringP("key", "k", "", "API key")
}
