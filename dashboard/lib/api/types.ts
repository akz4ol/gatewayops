export interface Trace {
  id: string;
  orgId: string;
  apiKeyId?: string;
  mcpServer: string;
  operation: string;
  status: 'success' | 'error' | 'timeout';
  startTime: string;
  endTime?: string;
  durationMs?: number;
  spans?: Span[];
  errorMessage?: string;
  cost?: number;
}

export interface Span {
  id: string;
  traceId: string;
  parentSpanId?: string;
  name: string;
  kind: string;
  status: string;
  startTime: string;
  endTime?: string;
  durationMs?: number;
  attributes?: Record<string, unknown>;
}

export interface TracePage {
  traces: Trace[];
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
}

export interface CostBreakdown {
  dimension: string;
  value: string;
  cost: number;
  requestCount: number;
}

export interface CostSummary {
  totalCost: number;
  periodStart: string;
  periodEnd: string;
  requestCount: number;
  byServer?: CostBreakdown[];
  byTeam?: CostBreakdown[];
  byTool?: CostBreakdown[];
}

export interface APIKey {
  id: string;
  name: string;
  keyPrefix: string;
  environment: string;
  permissions: string;
  rateLimitRpm: number;
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
}

export interface OverviewStats {
  totalRequests: number;
  totalCost: number;
  avgLatency: number;
  errorRate: number;
  activeServers: number;
  requestsChange: number;
  costChange: number;
  latencyChange: number;
  errorRateChange: number;
}

export interface TimeSeriesPoint {
  timestamp: string;
  value: number;
}

export interface Alert {
  id: string;
  ruleId: string;
  status: 'firing' | 'resolved';
  severity: 'info' | 'warning' | 'critical';
  message: string;
  startedAt: string;
  resolvedAt?: string;
}
