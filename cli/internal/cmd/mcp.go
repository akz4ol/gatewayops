package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/akz4ol/gatewayops/cli/internal/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Interact with MCP servers",
	Long:  `List and interact with MCP servers through the gateway.`,
}

var mcpToolsCmd = &cobra.Command{
	Use:   "tools [server]",
	Short: "List tools on an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())
		server := args[0]

		data, err := client.Post(fmt.Sprintf("/v1/mcp/%s/tools/list", server), nil)
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var result struct {
			Tools []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.Tools) == 0 {
			fmt.Println("No tools found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Tool", "Description"})
		table.SetBorder(false)
		table.SetColWidth(60)

		for _, t := range result.Tools {
			desc := t.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			table.Append([]string{t.Name, desc})
		}

		table.Render()
		fmt.Printf("\nTotal: %d tools on %s\n", len(result.Tools), server)
		return nil
	},
}

var mcpCallCmd = &cobra.Command{
	Use:   "call [server] [tool]",
	Short: "Call a tool on an MCP server",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())
		server := args[0]
		tool := args[1]

		argsJSON, _ := cmd.Flags().GetString("args")
		var toolArgs map[string]interface{}
		if argsJSON != "" {
			if err := json.Unmarshal([]byte(argsJSON), &toolArgs); err != nil {
				return fmt.Errorf("invalid JSON args: %w", err)
			}
		}

		body := map[string]interface{}{
			"tool":      tool,
			"arguments": toolArgs,
		}

		data, err := client.Post(fmt.Sprintf("/v1/mcp/%s/tools/call", server), body)
		if err != nil {
			return err
		}

		var result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		for _, c := range result.Content {
			if c.Type == "text" {
				fmt.Println(c.Text)
			}
		}

		return nil
	},
}

var mcpResourcesCmd = &cobra.Command{
	Use:   "resources [server]",
	Short: "List resources on an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())
		server := args[0]

		data, err := client.Post(fmt.Sprintf("/v1/mcp/%s/resources/list", server), nil)
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var result struct {
			Resources []struct {
				URI         string `json:"uri"`
				Name        string `json:"name"`
				Description string `json:"description"`
				MimeType    string `json:"mimeType"`
			} `json:"resources"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.Resources) == 0 {
			fmt.Println("No resources found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"URI", "Name", "Type"})
		table.SetBorder(false)

		for _, r := range result.Resources {
			table.Append([]string{r.URI, r.Name, r.MimeType})
		}

		table.Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpCallCmd)
	mcpCmd.AddCommand(mcpResourcesCmd)

	mcpCallCmd.Flags().StringP("args", "a", "", "Tool arguments as JSON")
}
