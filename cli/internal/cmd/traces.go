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

var tracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "Manage and view traces",
	Long:  `View and search API traces from your MCP Gateway.`,
}

var tracesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent traces",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		limit, _ := cmd.Flags().GetInt("limit")
		server, _ := cmd.Flags().GetString("server")

		path := fmt.Sprintf("/v1/traces?limit=%d", limit)
		if server != "" {
			path += "&mcp_server=" + server
		}

		data, err := client.Get(path)
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var result struct {
			Traces []struct {
				ID        string    `json:"id"`
				MCPServer string    `json:"mcp_server"`
				Operation string    `json:"operation"`
				Status    string    `json:"status"`
				Duration  int       `json:"duration_ms"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"traces"`
			Total int `json:"total"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.Traces) == 0 {
			fmt.Println("No traces found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Server", "Operation", "Status", "Duration", "Time"})
		table.SetBorder(false)

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		for _, t := range result.Traces {
			status := t.Status
			switch t.Status {
			case "success":
				status = green(t.Status)
			case "error":
				status = red(t.Status)
			default:
				status = yellow(t.Status)
			}

			table.Append([]string{
				t.ID[:8],
				t.MCPServer,
				t.Operation,
				status,
				fmt.Sprintf("%dms", t.Duration),
				t.CreatedAt.Format("15:04:05"),
			})
		}

		table.Render()
		fmt.Printf("\nTotal: %d traces\n", result.Total)
		return nil
	},
}

var tracesGetCmd = &cobra.Command{
	Use:   "get [trace-id]",
	Short: "Get trace details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		data, err := client.Get("/v1/traces/" + args[0])
		if err != nil {
			return err
		}

		if output == "json" {
			fmt.Println(string(data))
			return nil
		}

		var trace struct {
			ID        string                 `json:"id"`
			MCPServer string                 `json:"mcp_server"`
			Operation string                 `json:"operation"`
			Status    string                 `json:"status"`
			Duration  int                    `json:"duration_ms"`
			Request   map[string]interface{} `json:"request"`
			Response  map[string]interface{} `json:"response"`
			CreatedAt time.Time              `json:"created_at"`
		}
		if err := json.Unmarshal(data, &trace); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Printf("Trace: %s\n", trace.ID)
		fmt.Printf("Server: %s\n", trace.MCPServer)
		fmt.Printf("Operation: %s\n", trace.Operation)
		fmt.Printf("Status: %s\n", trace.Status)
		fmt.Printf("Duration: %dms\n", trace.Duration)
		fmt.Printf("Time: %s\n", trace.CreatedAt.Format(time.RFC3339))

		if trace.Request != nil {
			reqJSON, _ := json.MarshalIndent(trace.Request, "", "  ")
			fmt.Printf("\nRequest:\n%s\n", string(reqJSON))
		}

		if trace.Response != nil {
			respJSON, _ := json.MarshalIndent(trace.Response, "", "  ")
			fmt.Printf("\nResponse:\n%s\n", string(respJSON))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tracesCmd)
	tracesCmd.AddCommand(tracesListCmd)
	tracesCmd.AddCommand(tracesGetCmd)

	tracesListCmd.Flags().IntP("limit", "l", 20, "Number of traces to show")
	tracesListCmd.Flags().StringP("server", "s", "", "Filter by MCP server")
}
