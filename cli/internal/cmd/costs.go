package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/akz4ol/gatewayops/cli/internal/api"
)

var costsCmd = &cobra.Command{
	Use:   "costs",
	Short: "View cost information",
	Long:  `Commands for viewing usage costs and analytics.`,
}

var costsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "View cost summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		period, _ := cmd.Flags().GetString("period")
		groupBy, _ := cmd.Flags().GetString("group-by")

		path := fmt.Sprintf("/v1/costs/summary?period=%s", period)
		if groupBy != "" {
			path += "&group_by=" + groupBy
		}

		var result api.CostSummary
		if err := client.Get(path, &result); err != nil {
			return fmt.Errorf("failed to get cost summary: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		fmt.Printf("Cost Summary (%s)\n", period)
		fmt.Printf("Period: %s to %s\n",
			result.PeriodStart.Format("2006-01-02"),
			result.PeriodEnd.Format("2006-01-02"),
		)
		fmt.Printf("Total Cost: $%.2f\n", result.TotalCost)
		fmt.Printf("Total Requests: %d\n", result.RequestCount)

		if len(result.ByServer) > 0 {
			fmt.Println("\nBy Server:")
			printBreakdown(result.ByServer)
		}

		if len(result.ByTeam) > 0 {
			fmt.Println("\nBy Team:")
			printBreakdown(result.ByTeam)
		}

		if len(result.ByTool) > 0 {
			fmt.Println("\nBy Tool:")
			printBreakdown(result.ByTool)
		}

		return nil
	},
}

var costsByServerCmd = &cobra.Command{
	Use:   "by-server",
	Short: "View costs by MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		period, _ := cmd.Flags().GetString("period")

		path := fmt.Sprintf("/v1/costs/summary?period=%s&group_by=server", period)

		var result api.CostSummary
		if err := client.Get(path, &result); err != nil {
			return fmt.Errorf("failed to get costs: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.ByServer)
		}

		fmt.Printf("Costs by Server (%s)\n", period)
		fmt.Printf("Total: $%.2f\n\n", result.TotalCost)
		printBreakdown(result.ByServer)
		return nil
	},
}

var costsByTeamCmd = &cobra.Command{
	Use:   "by-team",
	Short: "View costs by team",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		period, _ := cmd.Flags().GetString("period")

		path := fmt.Sprintf("/v1/costs/summary?period=%s&group_by=team", period)

		var result api.CostSummary
		if err := client.Get(path, &result); err != nil {
			return fmt.Errorf("failed to get costs: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.ByTeam)
		}

		fmt.Printf("Costs by Team (%s)\n", period)
		fmt.Printf("Total: $%.2f\n\n", result.TotalCost)
		printBreakdown(result.ByTeam)
		return nil
	},
}

var costsByToolCmd = &cobra.Command{
	Use:   "by-tool",
	Short: "View costs by tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		period, _ := cmd.Flags().GetString("period")

		path := fmt.Sprintf("/v1/costs/summary?period=%s&group_by=tool", period)

		var result api.CostSummary
		if err := client.Get(path, &result); err != nil {
			return fmt.Errorf("failed to get costs: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result.ByTool)
		}

		fmt.Printf("Costs by Tool (%s)\n", period)
		fmt.Printf("Total: $%.2f\n\n", result.TotalCost)
		printBreakdown(result.ByTool)
		return nil
	},
}

func printBreakdown(breakdowns []api.CostBreakdown) {
	if len(breakdowns) == 0 {
		fmt.Println("  No data")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tCOST\tREQUESTS\t%")

	// Find max cost for percentage calculation
	var totalCost float64
	for _, b := range breakdowns {
		totalCost += b.Cost
	}

	for _, b := range breakdowns {
		pct := float64(0)
		if totalCost > 0 {
			pct = (b.Cost / totalCost) * 100
		}
		fmt.Fprintf(w, "  %s\t$%.2f\t%d\t%.1f%%\n",
			b.Value,
			b.Cost,
			b.RequestCount,
			pct,
		)
	}
	w.Flush()
}

func init() {
	costsSummaryCmd.Flags().Bool("json", false, "Output as JSON")
	costsSummaryCmd.Flags().StringP("period", "p", "month", "Time period (day, week, month)")
	costsSummaryCmd.Flags().StringP("group-by", "g", "", "Group by dimension (server, team, tool)")

	costsByServerCmd.Flags().Bool("json", false, "Output as JSON")
	costsByServerCmd.Flags().StringP("period", "p", "month", "Time period (day, week, month)")

	costsByTeamCmd.Flags().Bool("json", false, "Output as JSON")
	costsByTeamCmd.Flags().StringP("period", "p", "month", "Time period (day, week, month)")

	costsByToolCmd.Flags().Bool("json", false, "Output as JSON")
	costsByToolCmd.Flags().StringP("period", "p", "month", "Time period (day, week, month)")

	costsCmd.AddCommand(costsSummaryCmd)
	costsCmd.AddCommand(costsByServerCmd)
	costsCmd.AddCommand(costsByTeamCmd)
	costsCmd.AddCommand(costsByToolCmd)
}
