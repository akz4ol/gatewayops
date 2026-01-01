package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Commands for managing authentication with the GatewayOps API.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the GatewayOps API",
	Long: `Store your API key for authentication with the GatewayOps API.

You can also set the GATEWAYOPS_API_KEY environment variable instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")

		if key == "" {
			fmt.Print("Enter your GatewayOps API key: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			key = strings.TrimSpace(input)
		}

		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		// Validate the key format
		if !strings.HasPrefix(key, "gwo_") {
			return fmt.Errorf("invalid API key format. Keys should start with 'gwo_'")
		}

		// Save to config
		viper.Set("api_key", key)

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configPath := home + "/.gwo.yaml"
		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Successfully authenticated!")
		fmt.Printf("Config saved to %s\n", configPath)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configPath := home + "/.gwo.yaml"
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config: %w", err)
		}

		fmt.Println("Successfully logged out!")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		key := viper.GetString("api_key")
		if key == "" {
			key = os.Getenv("GATEWAYOPS_API_KEY")
		}

		if key == "" {
			fmt.Println("Not authenticated")
			fmt.Println("Run 'gwo auth login' to authenticate")
			return
		}

		// Show masked key
		if len(key) > 12 {
			fmt.Printf("Authenticated with key: %s...%s\n", key[:8], key[len(key)-4:])
		} else {
			fmt.Println("Authenticated")
		}
	},
}

func init() {
	loginCmd.Flags().StringP("key", "k", "", "API key to use")

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
}
