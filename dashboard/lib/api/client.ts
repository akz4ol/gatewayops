const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api';

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export class ApiClient {
  private baseUrl: string;
  private apiKey?: string;

  constructor(baseUrl: string = API_BASE, apiKey?: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  private async request<T>(
    method: string,
    path: string,
    options?: { body?: unknown; params?: Record<string, string> }
  ): Promise<T> {
    const url = new URL(path, this.baseUrl.startsWith('http') ? this.baseUrl : window.location.origin + this.baseUrl);
    if (options?.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        url.searchParams.set(key, value);
      });
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.apiKey) {
      headers['Authorization'] = `Bearer ${this.apiKey}`;
    }

    const response = await fetch(url.toString(), {
      method,
      headers,
      body: options?.body ? JSON.stringify(options.body) : undefined,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || 'Request failed');
    }

    return data;
  }

  get<T>(path: string, params?: Record<string, string>): Promise<T> {
    return this.request<T>('GET', path, { params });
  }

  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('POST', path, { body });
  }

  put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PUT', path, { body });
  }

  delete<T>(path: string): Promise<T> {
    return this.request<T>('DELETE', path);
  }

  // Dashboard metrics
  async getOverview() {
    return this.get<OverviewMetrics>('/metrics/overview');
  }

  async getRequestsChart() {
    return this.get<{ data: ChartDataPoint[] }>('/metrics/requests-chart');
  }

  async getTopServers() {
    return this.get<{ servers: ServerStats[] }>('/metrics/top-servers');
  }

  async getRecentTraces() {
    return this.get<{ traces: RecentTrace[] }>('/metrics/recent-traces');
  }

  // Traces
  async listTraces(params?: { limit?: number; offset?: number; server?: string; status?: string }) {
    const queryParams: Record<string, string> = {};
    if (params?.limit) queryParams.limit = String(params.limit);
    if (params?.offset) queryParams.offset = String(params.offset);
    if (params?.server) queryParams.server = params.server;
    if (params?.status) queryParams.status = params.status;
    return this.get<TracesResponse>('/traces', queryParams);
  }

  async getTrace(traceId: string) {
    return this.get<TraceDetail>(`/traces/${traceId}`);
  }

  async getTraceStats() {
    return this.get<TraceStats>('/traces/stats');
  }

  // Costs
  async getCostSummary(period?: string) {
    return this.get<CostSummaryResponse>('/costs/summary', period ? { period } : undefined);
  }

  async getCostsByServer() {
    return this.get<CostsByServerResponse>('/costs/by-server');
  }

  async getCostsByTeam() {
    return this.get<CostsByTeamResponse>('/costs/by-team');
  }

  async getDailyCosts() {
    return this.get<{ daily: DailyCost[] }>('/costs/daily');
  }

  // API Keys
  async listApiKeys() {
    return this.get<ApiKeysResponse>('/api-keys');
  }

  async createApiKey(data: CreateApiKeyRequest) {
    return this.post<ApiKeyCreated>('/api-keys', data);
  }

  async getApiKey(keyId: string) {
    return this.get<ApiKeyDetail>(`/api-keys/${keyId}`);
  }

  async deleteApiKey(keyId: string) {
    return this.delete<{ message: string }>(`/api-keys/${keyId}`);
  }

  async rotateApiKey(keyId: string) {
    return this.post<ApiKeyCreated>(`/api-keys/${keyId}/rotate`);
  }
}

// Types
export interface OverviewMetrics {
  total_requests: { value: number; change: number; period: string; formatted: string };
  total_cost: { value: number; change: number; period: string; formatted: string };
  avg_latency: { value: number; change: number; period: string; formatted: string; percentile: string };
  error_rate: { value: number; change: number; period: string; formatted: string };
}

export interface ChartDataPoint {
  date: string;
  requests: number;
  errors: number;
}

export interface ServerStats {
  name: string;
  requests: number;
  cost: number;
}

export interface RecentTrace {
  id: string;
  server: string;
  operation: string;
  tool: string;
  status: string;
  duration: number;
  time: string;
}

export interface TracesResponse {
  traces: Trace[];
  total: number;
  limit: number;
  offset: number;
}

export interface Trace {
  id: string;
  trace_id: string;
  span_id: string;
  org_id: string;
  mcp_server: string;
  operation: string;
  tool_name: string;
  status: string;
  status_code: number;
  duration_ms: number;
  request_size: number;
  response_size: number;
  cost: number;
  error_msg?: string;
  created_at: string;
}

export interface TraceDetail {
  trace: Trace;
  spans: TraceSpan[];
}

export interface TraceSpan {
  id: string;
  trace_id: string;
  span_id: string;
  parent_id?: string;
  name: string;
  kind: string;
  status: string;
  start_time: string;
  end_time: string;
  duration_ms: number;
}

export interface TraceStats {
  total_requests: number;
  success_count: number;
  error_count: number;
  avg_duration_ms: number;
  p50_duration_ms: number;
  p95_duration_ms: number;
  p99_duration_ms: number;
  total_cost: number;
  error_rate: number;
}

export interface CostSummaryResponse {
  total_cost: number;
  total_requests: number;
  avg_cost_per_req: number;
  period: string;
  start_date: string;
  end_date: string;
}

export interface CostsByServerResponse {
  total_cost: number;
  servers: CostByServer[];
}

export interface CostByServer {
  mcp_server: string;
  total_cost: number;
  total_requests: number;
  avg_cost_per_req: number;
  percentage: number;
}

export interface CostsByTeamResponse {
  total_cost: number;
  teams: CostByTeam[];
}

export interface CostByTeam {
  team_id: string;
  team_name: string;
  total_cost: number;
  total_requests: number;
  avg_cost_per_req: number;
  percentage: number;
}

export interface DailyCost {
  date: string;
  total_cost: number;
  total_requests: number;
}

export interface ApiKeysResponse {
  api_keys: ApiKey[];
  total: number;
}

export interface ApiKey {
  id: string;
  org_id: string;
  team_id?: string;
  name: string;
  key_prefix: string;
  environment: string;
  permissions: string[];
  rate_limit: number;
  last_used_at?: string;
  expires_at?: string;
  created_at: string;
  created_by: string;
  revoked: boolean;
  revoked_at?: string;
}

export interface CreateApiKeyRequest {
  name: string;
  environment?: string;
  permissions?: string[];
  rate_limit?: number;
  team_id?: string;
  expires_at?: string;
}

export interface ApiKeyCreated {
  id: string;
  org_id: string;
  name: string;
  key_prefix: string;
  environment: string;
  permissions: string[];
  rate_limit: number;
  created_at: string;
  raw_key: string;
}

export interface ApiKeyDetail {
  api_key: ApiKey;
  usage: {
    key_id: string;
    total_requests: number;
    total_cost: number;
    last_used_at: string;
  };
}

export const api = new ApiClient();
