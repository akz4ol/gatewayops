/**
 * Cost tracking and analytics types
 */

/**
 * Cost breakdown by dimension
 */
export interface CostBreakdown {
  dimension: string;
  value: string;
  cost: number;
  requestCount: number;
}

/**
 * Cost summary for a period (matches API response)
 */
export interface CostSummary {
  total_cost: number;
  total_requests: number;
  avg_cost_per_request: number;
  period: string;
  start_date: string;
  end_date: string;
  by_server?: CostBreakdown[];
  by_team?: CostBreakdown[];
  by_tool?: CostBreakdown[];
}

/**
 * Time period for cost queries
 */
export type CostPeriod = 'day' | 'week' | 'month';

/**
 * Grouping dimension for cost queries
 */
export type CostGroupBy = 'server' | 'team' | 'tool';

/**
 * Helper to get total cost from summary
 */
export function getTotalCost(summary: CostSummary): number {
  return summary.total_cost;
}

/**
 * Helper to get request count from summary
 */
export function getRequestCount(summary: CostSummary): number {
  return summary.total_requests;
}
