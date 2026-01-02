package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/akz4ol/gatewayops/cli/internal/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var costsCmd = &cobra.Command{
	Use:   "costs",
	Short: "View cost information",
	Long:  `View cost summaries and breakdowns for your MCP Gateway usage.`,
}

var costsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show cost summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(getBaseURL(), getAPIKey())

		period, _ := cmd.Flags().GetString("period")
		groupBy, _ := cmd.Flags().GetString("group-by")

		path := fmt.Sprintf("/v1/costs/summary?period=%s", period)
		if groupBy != "" {
			path += "&group_by=" + groupBy
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
			TotalCost         float64 `json:"total_cost"`
			TotalRequests     int     `json:"total_requests"`
			AvgCostPerRequest float64 `json:"avg_cost_per_request"`
			Period            string  `json:"period"`
			Breakdown         []struct {
				Name     string  `json:"name"`
				Cost     float64 `json:"cost"`
				Requests int     `json:"requests"`
			} `json:"breakdown"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Printf("Cost Summary (%s)\n", result.Period)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Total Cost:      $%.2f\n", result.TotalCost)
		fmt.Printf("Total Requests:  %d\n", result.TotalRequests)
		fmt.Printf("Avg Cost/Req:    $%.4f\n", result.AvgCostPerRequest)

		if len(result.Breakdown) > 0 {
			fmt.Printf("\nBreakdown by %s:\n", groupBy)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Cost", "Requests", "% of Total"})
			table.SetBorder(false)

			for _, b := range result.Breakdown {
				pct := 0.0
				if result.TotalCost > 0 {
					pct = (b.Cost / result.TotalCost) * 100
				}
				table.Append([]string{
					b.Name,
					fmt.Sprintf("$%.2f", b.Cost),
					fmt.Sprintf("%d", b.Requests),
					fmt.Sprintf("%.1f%%", pct),
				})
			}
			table.Render()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(costsCmd)
	costsCmd.AddCommand(costsSummaryCmd)

	costsSummaryCmd.Flags().StringP("period", "p", "month", "Time period (day, week, month)")
	costsSummaryCmd.Flags().StringP("group-by", "g", "", "Group by (server, team, tool)")
}
