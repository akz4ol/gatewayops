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
 * Cost summary for a period
 */
export interface CostSummary {
  totalCost: number;
  periodStart: Date;
  periodEnd: Date;
  requestCount: number;
  byServer?: CostBreakdown[];
  byTeam?: CostBreakdown[];
  byTool?: CostBreakdown[];
}

/**
 * Time period for cost queries
 */
export type CostPeriod = 'day' | 'week' | 'month';

/**
 * Grouping dimension for cost queries
 */
export type CostGroupBy = 'server' | 'team' | 'tool';
