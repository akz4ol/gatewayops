package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/akz4ol/gatewayops/cli/internal/api"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  `Commands for managing GatewayOps API keys.`,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")

		var result struct {
			Keys []api.APIKey `json:"keys"`
		}
		if err := client.Get("/v1/api-keys", &result); err != nil {
			return fmt.Errorf("failed to list keys: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Keys)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPREFIX\tENV\tPERMISSIONS\tRPM")
		for _, key := range result.Keys {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
				key.ID,
				key.Name,
				key.KeyPrefix,
				key.Environment,
				key.Permissions,
				key.RateLimitRPM,
			)
		}
		return w.Flush()
	},
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		env, _ := cmd.Flags().GetString("environment")
		permissions, _ := cmd.Flags().GetString("permissions")
		rpm, _ := cmd.Flags().GetInt("rpm")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]interface{}{
			"name":        name,
			"environment": env,
			"permissions": permissions,
		}
		if rpm > 0 {
			body["rateLimitRpm"] = rpm
		}

		var result struct {
			Key   api.APIKey `json:"key"`
			Token string     `json:"token"`
		}
		if err := client.Post("/v1/api-keys", body, &result); err != nil {
			return fmt.Errorf("failed to create key: %w", err)
		}

		fmt.Printf("Created API key: %s\n", result.Key.Name)
		fmt.Printf("ID: %s\n", result.Key.ID)
		fmt.Println()
		fmt.Println("Your API key (save this, it won't be shown again):")
		fmt.Printf("  %s\n", result.Token)
		return nil
	},
}

var keysDeleteCmd = &cobra.Command{
	Use:   "delete [key-id]",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]

		if err := client.Delete("/v1/api-keys/"+keyID, nil); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}

		fmt.Printf("Deleted API key: %s\n", keyID)
		return nil
	},
}

var keysRevokeCmd = &cobra.Command{
	Use:   "revoke [key-id]",
	Short: "Revoke an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]

		if err := client.Post("/v1/api-keys/"+keyID+"/revoke", nil, nil); err != nil {
			return fmt.Errorf("failed to revoke key: %w", err)
		}

		fmt.Printf("Revoked API key: %s\n", keyID)
		return nil
	},
}

func init() {
	keysListCmd.Flags().Bool("json", false, "Output as JSON")

	keysCreateCmd.Flags().StringP("name", "n", "", "Name for the API key (required)")
	keysCreateCmd.Flags().StringP("environment", "e", "production", "Environment (production or sandbox)")
	keysCreateCmd.Flags().StringP("permissions", "p", "full", "Permissions (full, read, write)")
	keysCreateCmd.Flags().Int("rpm", 0, "Rate limit (requests per minute)")

	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysCreateCmd)
	keysCmd.AddCommand(keysDeleteCmd)
	keysCmd.AddCommand(keysRevokeCmd)
}
