package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/akz4ol/gatewayops/cli/internal/api"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server operations",
	Long:  `Commands for interacting with MCP servers through the gateway.`,
}

// Tools commands
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tool operations",
}

var toolsListCmd = &cobra.Command{
	Use:   "list [server]",
	Short: "List available tools",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		outputJSON, _ := cmd.Flags().GetBool("json")

		var result struct {
			Tools []api.ToolDefinition `json:"tools"`
		}
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/tools/list", server), nil, &result); err != nil {
			return fmt.Errorf("failed to list tools: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Tools)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		for _, tool := range result.Tools {
			desc := tool.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\n", tool.Name, desc)
		}
		return w.Flush()
	},
}

var toolsCallCmd = &cobra.Command{
	Use:   "call [server] [tool]",
	Short: "Call a tool",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		tool := args[1]
		outputJSON, _ := cmd.Flags().GetBool("json")
		argsJSON, _ := cmd.Flags().GetString("args")

		var toolArgs map[string]interface{}
		if argsJSON != "" {
			if err := json.Unmarshal([]byte(argsJSON), &toolArgs); err != nil {
				return fmt.Errorf("invalid --args JSON: %w", err)
			}
		}

		body := map[string]interface{}{
			"tool":      tool,
			"arguments": toolArgs,
		}

		var result api.ToolCallResult
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/tools/call", server), body, &result); err != nil {
			return fmt.Errorf("tool call failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if result.IsError {
			fmt.Fprintf(os.Stderr, "Error: %v\n", result.Content)
			os.Exit(1)
		}

		// Pretty print the content
		content, err := json.MarshalIndent(result.Content, "", "  ")
		if err != nil {
			fmt.Println(result.Content)
		} else {
			fmt.Println(string(content))
		}

		if result.TraceID != "" {
			fmt.Fprintf(os.Stderr, "\nTrace ID: %s\n", result.TraceID)
		}
		if result.DurationMs > 0 {
			fmt.Fprintf(os.Stderr, "Duration: %dms\n", result.DurationMs)
		}
		if result.Cost > 0 {
			fmt.Fprintf(os.Stderr, "Cost: $%.4f\n", result.Cost)
		}

		return nil
	},
}

// Resources commands
var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Resource operations",
}

var resourcesListCmd = &cobra.Command{
	Use:   "list [server]",
	Short: "List available resources",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		outputJSON, _ := cmd.Flags().GetBool("json")

		var result struct {
			Resources []api.Resource `json:"resources"`
		}
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/resources/list", server), nil, &result); err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Resources)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "URI\tNAME\tTYPE")
		for _, res := range result.Resources {
			fmt.Fprintf(w, "%s\t%s\t%s\n", res.URI, res.Name, res.MimeType)
		}
		return w.Flush()
	},
}

var resourcesReadCmd = &cobra.Command{
	Use:   "read [server] [uri]",
	Short: "Read a resource",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		uri := args[1]
		outputJSON, _ := cmd.Flags().GetBool("json")

		body := map[string]interface{}{
			"uri": uri,
		}

		var result api.ResourceContent
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/resources/read", server), body, &result); err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if result.Text != "" {
			fmt.Println(result.Text)
		} else if result.Blob != "" {
			fmt.Printf("[Binary content: %d bytes]\n", len(result.Blob))
		}

		return nil
	},
}

// Prompts commands
var promptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Prompt operations",
}

var promptsListCmd = &cobra.Command{
	Use:   "list [server]",
	Short: "List available prompts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		outputJSON, _ := cmd.Flags().GetBool("json")

		var result struct {
			Prompts []api.Prompt `json:"prompts"`
		}
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/prompts/list", server), nil, &result); err != nil {
			return fmt.Errorf("failed to list prompts: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Prompts)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		for _, p := range result.Prompts {
			desc := p.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\n", p.Name, desc)
		}
		return w.Flush()
	},
}

var promptsGetCmd = &cobra.Command{
	Use:   "get [server] [name]",
	Short: "Get a prompt",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		name := args[1]
		outputJSON, _ := cmd.Flags().GetBool("json")
		argsJSON, _ := cmd.Flags().GetString("args")

		var promptArgs map[string]interface{}
		if argsJSON != "" {
			if err := json.Unmarshal([]byte(argsJSON), &promptArgs); err != nil {
				return fmt.Errorf("invalid --args JSON: %w", err)
			}
		}

		body := map[string]interface{}{
			"name":      name,
			"arguments": promptArgs,
		}

		var result struct {
			Messages []api.PromptMessage `json:"messages"`
		}
		if err := client.Post(fmt.Sprintf("/v1/mcp/%s/prompts/get", server), body, &result); err != nil {
			return fmt.Errorf("failed to get prompt: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Messages)
		}

		for _, msg := range result.Messages {
			fmt.Printf("[%s]\n", msg.Role)
			content, err := json.MarshalIndent(msg.Content, "", "  ")
			if err != nil {
				fmt.Println(msg.Content)
			} else {
				fmt.Println(string(content))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	// Tools flags
	toolsListCmd.Flags().Bool("json", false, "Output as JSON")
	toolsCallCmd.Flags().Bool("json", false, "Output as JSON")
	toolsCallCmd.Flags().StringP("args", "a", "", "Tool arguments as JSON")

	// Resources flags
	resourcesListCmd.Flags().Bool("json", false, "Output as JSON")
	resourcesReadCmd.Flags().Bool("json", false, "Output as JSON")

	// Prompts flags
	promptsListCmd.Flags().Bool("json", false, "Output as JSON")
	promptsGetCmd.Flags().Bool("json", false, "Output as JSON")
	promptsGetCmd.Flags().StringP("args", "a", "", "Prompt arguments as JSON")

	// Add tool commands
	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsCallCmd)

	// Add resource commands
	resourcesCmd.AddCommand(resourcesListCmd)
	resourcesCmd.AddCommand(resourcesReadCmd)

	// Add prompt commands
	promptsCmd.AddCommand(promptsListCmd)
	promptsCmd.AddCommand(promptsGetCmd)

	// Add to mcp
	mcpCmd.AddCommand(toolsCmd)
	mcpCmd.AddCommand(resourcesCmd)
	mcpCmd.AddCommand(promptsCmd)
}
