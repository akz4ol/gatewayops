package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/akz4ol/gatewayops/cli/internal/api"
)

var tracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "View and search traces",
	Long:  `Commands for viewing distributed traces.`,
}

var tracesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List traces",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		server, _ := cmd.Flags().GetString("server")
		operation, _ := cmd.Flags().GetString("operation")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		path := fmt.Sprintf("/v1/traces?limit=%d", limit)
		if server != "" {
			path += "&mcp_server=" + server
		}
		if operation != "" {
			path += "&operation=" + operation
		}
		if status != "" {
			path += "&status=" + status
		}

		var result api.TracePage
		if err := client.Get(path, &result); err != nil {
			return fmt.Errorf("failed to list traces: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if len(result.Traces) == 0 {
			fmt.Println("No traces found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSERVER\tOPERATION\tSTATUS\tDURATION\tTIME")
		for _, trace := range result.Traces {
			duration := "-"
			if trace.DurationMs > 0 {
				duration = fmt.Sprintf("%dms", trace.DurationMs)
			}
			age := formatAge(trace.StartTime)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				truncate(trace.ID, 12),
				trace.MCPServer,
				trace.Operation,
				colorStatus(trace.Status),
				duration,
				age,
			)
		}
		if err := w.Flush(); err != nil {
			return err
		}

		if result.HasMore {
			fmt.Printf("\n(showing %d of %d traces)\n", len(result.Traces), result.Total)
		}

		return nil
	},
}

var tracesGetCmd = &cobra.Command{
	Use:   "get [trace-id]",
	Short: "Get trace details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		traceID := args[0]
		outputJSON, _ := cmd.Flags().GetBool("json")

		var trace api.Trace
		if err := client.Get("/v1/traces/"+traceID, &trace); err != nil {
			return fmt.Errorf("failed to get trace: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(trace)
		}

		fmt.Printf("Trace: %s\n", trace.ID)
		fmt.Printf("Server: %s\n", trace.MCPServer)
		fmt.Printf("Operation: %s\n", trace.Operation)
		fmt.Printf("Status: %s\n", colorStatus(trace.Status))
		fmt.Printf("Started: %s\n", trace.StartTime.Format(time.RFC3339))
		if trace.DurationMs > 0 {
			fmt.Printf("Duration: %dms\n", trace.DurationMs)
		}
		if trace.Cost > 0 {
			fmt.Printf("Cost: $%.4f\n", trace.Cost)
		}
		if trace.ErrorMessage != "" {
			fmt.Printf("Error: %s\n", trace.ErrorMessage)
		}

		if len(trace.Spans) > 0 {
			fmt.Println("\nSpans:")
			printSpanTree(trace.Spans, "", "")
		}

		return nil
	},
}

func printSpanTree(spans []api.Span, parentID string, indent string) {
	for _, span := range spans {
		if span.ParentSpanID != parentID {
			continue
		}

		statusIcon := "+"
		if span.Status == "error" {
			statusIcon = "x"
		}

		duration := "-"
		if span.DurationMs > 0 {
			duration = fmt.Sprintf("%dms", span.DurationMs)
		}

		fmt.Printf("%s[%s] %s (%s)\n", indent, statusIcon, span.Name, duration)

		// Print child spans
		printSpanTree(spans, span.ID, indent+"  ")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func colorStatus(status string) string {
	switch strings.ToLower(status) {
	case "success":
		return "\033[32m" + status + "\033[0m" // Green
	case "error":
		return "\033[31m" + status + "\033[0m" // Red
	case "timeout":
		return "\033[33m" + status + "\033[0m" // Yellow
	default:
		return status
	}
}

func init() {
	tracesListCmd.Flags().Bool("json", false, "Output as JSON")
	tracesListCmd.Flags().StringP("server", "s", "", "Filter by MCP server")
	tracesListCmd.Flags().StringP("operation", "o", "", "Filter by operation")
	tracesListCmd.Flags().String("status", "", "Filter by status (success, error, timeout)")
	tracesListCmd.Flags().IntP("limit", "n", 20, "Maximum number of traces to show")

	tracesGetCmd.Flags().Bool("json", false, "Output as JSON")

	tracesCmd.AddCommand(tracesListCmd)
	tracesCmd.AddCommand(tracesGetCmd)
}
