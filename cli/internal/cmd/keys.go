package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/akz4ol/gatewayops/cli/internal/api"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  `Create, list, and revoke API keys for your organization.`,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		data, err := client.Get("/v1/api-keys")
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var result struct {
			APIKeys []struct {
				ID          string     `json:"id"`
				Name        string     `json:"name"`
				KeyPrefix   string     `json:"key_prefix"`
				Environment string     `json:"environment"`
				RateLimit   int        `json:"rate_limit"`
				LastUsedAt  *time.Time `json:"last_used_at"`
				Revoked     bool       `json:"revoked"`
				CreatedAt   time.Time  `json:"created_at"`
			} `json:"api_keys"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.APIKeys) == 0 {
			fmt.Println("No API keys found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Prefix", "Environment", "Rate Limit", "Last Used", "Status"})
		table.SetBorder(false)

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		for _, k := range result.APIKeys {
			status := green("Active")
			if k.Revoked {
				status = red("Revoked")
			}

			lastUsed := "Never"
			if k.LastUsedAt != nil {
				lastUsed = k.LastUsedAt.Format("Jan 02 15:04")
			}

			table.Append([]string{
				k.Name,
				k.KeyPrefix,
				k.Environment,
				fmt.Sprintf("%d/min", k.RateLimit),
				lastUsed,
				status,
			})
		}

		table.Render()
		return nil
	},
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		name, _ := cmd.Flags().GetString("name")
		env, _ := cmd.Flags().GetString("environment")
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")

		body := map[string]interface{}{
			"name":        name,
			"environment": env,
			"rate_limit":  rateLimit,
		}

		data, err := client.Post("/v1/api-keys", body)
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var result struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			RawKey string `json:"raw_key"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Printf("API Key created successfully!\n\n")
		fmt.Printf("Name: %s\n", result.Name)
		fmt.Printf("ID: %s\n", result.ID)
		fmt.Printf("\n")
		color.New(color.FgYellow, color.Bold).Println("Save this key - it will not be shown again:")
		color.New(color.FgGreen).Println(result.RawKey)
		return nil
	},
}

var keysRevokeCmd = &cobra.Command{
	Use:   "revoke [key-id]",
	Short: "Revoke an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		_, err := client.Delete("/v1/api-keys/" + args[0])
		if err != nil {
			return err
		}

		fmt.Println("API key revoked successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysCreateCmd)
	keysCmd.AddCommand(keysRevokeCmd)

	keysCreateCmd.Flags().StringP("name", "n", "", "Name for the API key (required)")
	keysCreateCmd.Flags().StringP("environment", "e", "development", "Environment (development, staging, production)")
	keysCreateCmd.Flags().IntP("rate-limit", "r", 100, "Rate limit (requests per minute)")
	keysCreateCmd.MarkFlagRequired("name")
}
