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

  // Alert Rules
  async listAlertRules() {
    return this.get<AlertRulesResponse>('/alerts/rules');
  }

  async getAlertRule(ruleId: string) {
    return this.get<AlertRule>(`/alerts/rules/${ruleId}`);
  }

  async createAlertRule(data: CreateAlertRuleRequest) {
    return this.post<AlertRule>('/alerts/rules', data);
  }

  async updateAlertRule(ruleId: string, data: UpdateAlertRuleRequest) {
    return this.put<AlertRule>(`/alerts/rules/${ruleId}`, data);
  }

  async deleteAlertRule(ruleId: string) {
    return this.delete<{ status: string }>(`/alerts/rules/${ruleId}`);
  }

  // Alert Channels
  async listAlertChannels() {
    return this.get<AlertChannelsResponse>('/alerts/channels');
  }

  async getAlertChannel(channelId: string) {
    return this.get<AlertChannel>(`/alerts/channels/${channelId}`);
  }

  async createAlertChannel(data: CreateAlertChannelRequest) {
    return this.post<AlertChannel>('/alerts/channels', data);
  }

  async updateAlertChannel(channelId: string, data: UpdateAlertChannelRequest) {
    return this.put<AlertChannel>(`/alerts/channels/${channelId}`, data);
  }

  async deleteAlertChannel(channelId: string) {
    return this.delete<{ status: string }>(`/alerts/channels/${channelId}`);
  }

  async testAlertChannel(channelId: string) {
    return this.post<{ status: string; message: string }>(`/alerts/channels/${channelId}/test`);
  }

  // Alerts
  async listAlerts(params?: AlertsFilterParams) {
    const queryParams: Record<string, string> = {};
    if (params?.rule_id) queryParams.rule_id = params.rule_id;
    if (params?.statuses) queryParams.statuses = params.statuses.join(',');
    if (params?.severities) queryParams.severities = params.severities.join(',');
    if (params?.limit) queryParams.limit = String(params.limit);
    if (params?.offset) queryParams.offset = String(params.offset);
    return this.get<AlertsResponse>('/alerts', queryParams);
  }

  async getActiveAlerts() {
    return this.get<ActiveAlertsResponse>('/alerts/active');
  }

  async acknowledgeAlert(alertId: string) {
    return this.post<Alert>(`/alerts/${alertId}/ack`);
  }

  async resolveAlert(alertId: string) {
    return this.post<Alert>(`/alerts/${alertId}/resolve`);
  }

  async triggerTestAlert(data?: { metric?: string; value?: number }) {
    return this.post<Alert>('/alerts/test', data);
  }

  // Safety Policies
  async listSafetyPolicies() {
    return this.get<SafetyPoliciesResponse>('/safety-policies');
  }

  async getSafetyPolicy(policyId: string) {
    return this.get<SafetyPolicy>(`/safety-policies/${policyId}`);
  }

  async createSafetyPolicy(data: CreateSafetyPolicyRequest) {
    return this.post<SafetyPolicy>('/safety-policies', data);
  }

  async updateSafetyPolicy(policyId: string, data: UpdateSafetyPolicyRequest) {
    return this.put<SafetyPolicy>(`/safety-policies/${policyId}`, data);
  }

  async deleteSafetyPolicy(policyId: string) {
    return this.delete<void>(`/safety-policies/${policyId}`);
  }

  // Safety Detections
  async listDetections(params?: DetectionsFilterParams) {
    const queryParams: Record<string, string> = {};
    if (params?.mcp_server) queryParams.mcp_server = params.mcp_server;
    if (params?.limit) queryParams.limit = String(params.limit);
    if (params?.offset) queryParams.offset = String(params.offset);
    return this.get<DetectionsResponse>('/safety/detections', queryParams);
  }

  async getSafetySummary() {
    return this.get<SafetySummary>('/safety/summary');
  }

  async testSafetyInput(data: SafetyTestRequest) {
    return this.post<SafetyTestResponse>('/safety/test', data);
  }

  // Users
  async listUsers(params?: { limit?: number; offset?: number }) {
    const queryParams: Record<string, string> = {};
    if (params?.limit) queryParams.limit = String(params.limit);
    if (params?.offset) queryParams.offset = String(params.offset);
    return this.get<UsersResponse>('/users', queryParams);
  }

  async getUser(userId: string) {
    return this.get<User>(`/users/${userId}`);
  }

  // Invites
  async listInvites() {
    return this.get<InvitesResponse>('/users/invites');
  }

  async createInvite(data: CreateInviteRequest) {
    return this.post<Invite>('/users/invites', data);
  }

  async cancelInvite(inviteId: string) {
    return this.delete<{ status: string }>(`/users/invites/${inviteId}`);
  }

  async resendInvite(inviteId: string) {
    return this.post<{ status: string; invite: Invite }>(`/users/invites/${inviteId}/resend`);
  }

  // Roles (via RBAC)
  async listRoles() {
    return this.get<RolesResponse>('/rbac/roles');
  }

  // Settings
  async getSettings() {
    return this.get<OrgSettings>('/settings');
  }

  async updateSettings(data: UpdateSettingsRequest) {
    return this.put<OrgSettings>('/settings', data);
  }

  // SSO Providers
  async listSSOProviders(includeDisabled?: boolean) {
    const params: Record<string, string> = {};
    if (includeDisabled) params.include_disabled = 'true';
    return this.get<SSOProvidersResponse>('/sso/providers', params);
  }

  async getSSOProvider(providerId: string) {
    return this.get<SSOProvider>(`/sso/providers/${providerId}`);
  }

  async createSSOProvider(data: CreateSSOProviderRequest) {
    return this.post<SSOProvider>('/sso/providers', data);
  }

  async updateSSOProvider(providerId: string, data: UpdateSSOProviderRequest) {
    return this.put<SSOProvider>(`/sso/providers/${providerId}`, data);
  }

  async deleteSSOProvider(providerId: string) {
    return this.delete<{ status: string }>(`/sso/providers/${providerId}`);
  }

  async testSSOProvider(providerId: string) {
    return this.post<{ status: string; message: string }>(`/sso/providers/${providerId}/test`);
  }

  async getSupportedSSOProviders() {
    return this.get<SupportedSSOProvidersResponse>('/sso/providers/supported');
  }

  // Telemetry Configs
  async listTelemetryConfigs() {
    return this.get<TelemetryConfigsResponse>('/telemetry/configs');
  }

  async getTelemetryConfig(configId: string) {
    return this.get<TelemetryConfig>(`/telemetry/configs/${configId}`);
  }

  async createTelemetryConfig(data: CreateTelemetryConfigRequest) {
    return this.post<TelemetryConfig>('/telemetry/configs', data);
  }

  async updateTelemetryConfig(configId: string, data: UpdateTelemetryConfigRequest) {
    return this.put<TelemetryConfig>(`/telemetry/configs/${configId}`, data);
  }

  async deleteTelemetryConfig(configId: string) {
    return this.delete<{ status: string }>(`/telemetry/configs/${configId}`);
  }

  async testTelemetryConfig(configId: string) {
    return this.post<TelemetryTestResult>(`/telemetry/configs/${configId}/test`);
  }

  async getSupportedExporters() {
    return this.get<SupportedExportersResponse>('/telemetry/exporters');
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

// Alert Types
export type AlertSeverity = 'info' | 'warning' | 'critical';
export type AlertStatus = 'firing' | 'resolved' | 'acknowledged';
export type AlertMetric = 'error_rate' | 'latency_p50' | 'latency_p95' | 'latency_p99' | 'request_rate' | 'cost_per_hour' | 'cost_per_day' | 'rate_limit_hit' | 'injection_detected';
export type AlertCondition = 'gt' | 'lt' | 'gte' | 'lte' | 'eq' | 'neq';
export type AlertChannelType = 'slack' | 'pagerduty' | 'opsgenie' | 'webhook' | 'email' | 'teams';

export interface AlertFilters {
  mcp_servers?: string[];
  teams?: string[];
  environments?: string[];
}

export interface AlertRule {
  id: string;
  org_id: string;
  name: string;
  description?: string;
  metric: AlertMetric;
  condition: AlertCondition;
  threshold: number;
  window_minutes: number;
  severity: AlertSeverity;
  channels: string[];
  filters?: AlertFilters;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface AlertRulesResponse {
  rules: AlertRule[];
  total: number;
}

export interface CreateAlertRuleRequest {
  name: string;
  description?: string;
  metric: AlertMetric;
  condition: AlertCondition;
  threshold: number;
  window_minutes?: number;
  severity?: AlertSeverity;
  channels?: string[];
  filters?: AlertFilters;
  enabled?: boolean;
}

export interface UpdateAlertRuleRequest {
  name?: string;
  description?: string;
  metric?: AlertMetric;
  condition?: AlertCondition;
  threshold?: number;
  window_minutes?: number;
  severity?: AlertSeverity;
  channels?: string[];
  filters?: AlertFilters;
  enabled?: boolean;
}

export interface AlertChannel {
  id: string;
  org_id: string;
  name: string;
  type: AlertChannelType;
  config: Record<string, unknown>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface AlertChannelsResponse {
  channels: AlertChannel[];
  total: number;
}

export interface CreateAlertChannelRequest {
  name: string;
  type: AlertChannelType;
  config: Record<string, unknown>;
  enabled?: boolean;
}

export interface UpdateAlertChannelRequest {
  name?: string;
  type?: AlertChannelType;
  config?: Record<string, unknown>;
  enabled?: boolean;
}

export interface Alert {
  id: string;
  org_id: string;
  rule_id: string;
  status: AlertStatus;
  severity: AlertSeverity;
  message: string;
  value: number;
  threshold: number;
  labels?: Record<string, string>;
  started_at: string;
  resolved_at?: string;
  acked_at?: string;
  acked_by?: string;
}

export interface AlertsResponse {
  alerts: Alert[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface ActiveAlertsResponse {
  alerts: Alert[];
  total: number;
}

export interface AlertsFilterParams {
  rule_id?: string;
  statuses?: AlertStatus[];
  severities?: AlertSeverity[];
  start_time?: string;
  end_time?: string;
  limit?: number;
  offset?: number;
}

// Safety Types
export type SafetyMode = 'block' | 'warn' | 'log';
export type SafetySensitivity = 'strict' | 'moderate' | 'permissive';
export type DetectionSeverity = 'low' | 'medium' | 'high' | 'critical';
export type DetectionType = 'prompt_injection' | 'pii' | 'secret' | 'malicious';

export interface SafetyPatterns {
  block?: string[];
  allow?: string[];
}

export interface SafetyPolicy {
  id: string;
  org_id: string;
  name: string;
  description?: string;
  sensitivity: SafetySensitivity;
  mode: SafetyMode;
  patterns: SafetyPatterns;
  mcp_servers?: string[];
  enabled: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface SafetyPoliciesResponse {
  policies: SafetyPolicy[];
}

export interface CreateSafetyPolicyRequest {
  name: string;
  description?: string;
  sensitivity?: SafetySensitivity;
  mode?: SafetyMode;
  patterns?: SafetyPatterns;
  mcp_servers?: string[];
  enabled?: boolean;
}

export interface UpdateSafetyPolicyRequest {
  name?: string;
  description?: string;
  sensitivity?: SafetySensitivity;
  mode?: SafetyMode;
  patterns?: SafetyPatterns;
  mcp_servers?: string[];
  enabled?: boolean;
}

export interface InjectionDetection {
  id: string;
  org_id: string;
  trace_id?: string;
  span_id?: string;
  policy_id?: string;
  type: DetectionType;
  severity: DetectionSeverity;
  pattern_matched?: string;
  input: string;
  action_taken: SafetyMode;
  mcp_server?: string;
  tool_name?: string;
  api_key_id?: string;
  ip_address?: string;
  created_at: string;
}

export interface DetectionsResponse {
  detections: InjectionDetection[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface DetectionsFilterParams {
  mcp_server?: string;
  limit?: number;
  offset?: number;
}

export interface DetectionResult {
  detected: boolean;
  type?: DetectionType;
  severity?: DetectionSeverity;
  pattern_matched?: string;
  confidence?: number;
  action: SafetyMode;
  message?: string;
}

export interface SafetyTestRequest {
  input: string;
  policy_id?: string;
}

export interface SafetyTestResponse {
  result: DetectionResult;
  policy_id?: string;
}

export interface PatternCount {
  pattern: string;
  count: number;
}

export interface SafetySummary {
  total_detections: number;
  by_type: Record<string, number>;
  by_severity: Record<string, number>;
  by_action: Record<string, number>;
  top_patterns: PatternCount[];
  period: string;
}

// User & Team Types
export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  status: string;
  role: string;
  last_active_at?: string;
  created_at: string;
}

export interface UsersResponse {
  users: User[];
  total: number;
  limit: number;
  offset: number;
}

export interface Invite {
  id: string;
  org_id: string;
  email: string;
  role: string;
  invited_by: string;
  inviter_name: string;
  status: string;
  created_at: string;
  expires_at: string;
}

export interface InvitesResponse {
  invites: Invite[];
  total: number;
}

export interface CreateInviteRequest {
  email: string;
  role?: string;
}

export interface Role {
  id: string;
  org_id?: string;
  name: string;
  description?: string;
  permissions: string[];
  is_builtin: boolean;
  created_at: string;
  updated_at: string;
}

export interface RolesResponse {
  roles: Role[];
  total: number;
}

// Settings Types
export interface RateLimitConfig {
  production_rpm: number;
  sandbox_rpm: number;
}

export interface OrgSettings {
  id: string;
  org_id: string;
  org_name: string;
  billing_email: string;
  rate_limits: RateLimitConfig;
  updated_at: string;
}

export interface UpdateSettingsRequest {
  org_name?: string;
  billing_email?: string;
  rate_limits?: Partial<RateLimitConfig>;
}

// SSO Provider Types
export type SSOProviderType = 'okta' | 'azure_ad' | 'google' | 'onelogin' | 'auth0' | 'oidc';

export interface SSOProvider {
  id: string;
  org_id: string;
  type: SSOProviderType;
  name: string;
  issuer_url: string;
  client_id: string;
  authorization_url?: string;
  token_url?: string;
  userinfo_url?: string;
  scopes?: string[];
  claim_mappings?: Record<string, string>;
  group_mappings?: Record<string, string>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface SSOProvidersResponse {
  providers: SSOProvider[];
  total: number;
}

export interface CreateSSOProviderRequest {
  type: SSOProviderType;
  name: string;
  issuer_url: string;
  client_id: string;
  client_secret: string;
  scopes?: string[];
  claim_mappings?: Record<string, string>;
}

export interface UpdateSSOProviderRequest {
  name?: string;
  issuer_url?: string;
  client_id?: string;
  client_secret?: string;
  scopes?: string[];
  claim_mappings?: Record<string, string>;
  enabled?: boolean;
}

export interface SupportedSSOProvider {
  type: string;
  name: string;
  description: string;
  logo_url?: string;
}

export interface SupportedSSOProvidersResponse {
  supported_providers: SupportedSSOProvider[];
  total: number;
}

// Telemetry Config Types
export type TelemetryExporterType = 'otlp' | 'jaeger' | 'zipkin' | 'datadog' | 'newrelic';
export type TelemetryProtocol = 'grpc' | 'http';

export interface TelemetryConfig {
  id: string;
  org_id: string;
  name: string;
  exporter_type: TelemetryExporterType;
  endpoint: string;
  protocol: TelemetryProtocol;
  headers?: Record<string, string>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface TelemetryConfigsResponse {
  configs: TelemetryConfig[];
  total: number;
}

export interface CreateTelemetryConfigRequest {
  name: string;
  exporter_type?: TelemetryExporterType;
  endpoint: string;
  protocol?: TelemetryProtocol;
  headers?: Record<string, string>;
  enabled?: boolean;
}

export interface UpdateTelemetryConfigRequest {
  name?: string;
  exporter_type?: TelemetryExporterType;
  endpoint?: string;
  protocol?: TelemetryProtocol;
  headers?: Record<string, string>;
  enabled?: boolean;
}

export interface TelemetryTestResult {
  success: boolean;
  message: string;
  config_id: string;
  latency_ms?: number;
}

export interface SupportedExporter {
  type: TelemetryExporterType;
  name: string;
  description: string;
  protocols: TelemetryProtocol[];
}

export interface SupportedExportersResponse {
  exporters: SupportedExporter[];
}

export const api = new ApiClient();
