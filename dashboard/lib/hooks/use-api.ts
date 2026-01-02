import useSWR, { SWRConfiguration } from 'swr';
import useSWRMutation from 'swr/mutation';
import {
  api,
  OverviewMetrics,
  ChartDataPoint,
  ServerStats,
  RecentTrace,
  TracesResponse,
  TraceDetail,
  TraceStats,
  CostSummaryResponse,
  CostsByServerResponse,
  CostsByTeamResponse,
  DailyCost,
  ApiKeysResponse,
  ApiKeyDetail,
  ApiKeyCreated,
  CreateApiKeyRequest,
  AlertRulesResponse,
  AlertRule,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
  AlertChannelsResponse,
  AlertChannel,
  CreateAlertChannelRequest,
  UpdateAlertChannelRequest,
  AlertsResponse,
  ActiveAlertsResponse,
  Alert,
  AlertsFilterParams,
  SafetyPoliciesResponse,
  SafetyPolicy,
  CreateSafetyPolicyRequest,
  UpdateSafetyPolicyRequest,
  DetectionsResponse,
  DetectionsFilterParams,
  SafetySummary,
  SafetyTestRequest,
  SafetyTestResponse,
  UsersResponse,
  User,
  InvitesResponse,
  Invite,
  CreateInviteRequest,
  RolesResponse,
  OrgSettings,
  UpdateSettingsRequest,
  SSOProvidersResponse,
  SSOProvider,
  CreateSSOProviderRequest,
  UpdateSSOProviderRequest,
  SupportedSSOProvidersResponse,
  TelemetryConfigsResponse,
  TelemetryConfig,
  CreateTelemetryConfigRequest,
  UpdateTelemetryConfigRequest,
  TelemetryTestResult,
  SupportedExportersResponse,
} from '@/lib/api/client';

// Fetcher function for SWR
const fetcher = <T>(fn: () => Promise<T>) => fn();

// Default SWR options
const defaultOptions: SWRConfiguration = {
  revalidateOnFocus: false,
  revalidateOnReconnect: true,
  dedupingInterval: 5000,
};

// Dashboard Overview Hooks
export function useOverview(options?: SWRConfiguration) {
  return useSWR<OverviewMetrics>(
    'overview',
    () => fetcher(() => api.getOverview()),
    { ...defaultOptions, ...options }
  );
}

export function useRequestsChart(options?: SWRConfiguration) {
  return useSWR<{ data: ChartDataPoint[] }>(
    'requests-chart',
    () => fetcher(() => api.getRequestsChart()),
    { ...defaultOptions, ...options }
  );
}

export function useTopServers(options?: SWRConfiguration) {
  return useSWR<{ servers: ServerStats[] }>(
    'top-servers',
    () => fetcher(() => api.getTopServers()),
    { ...defaultOptions, ...options }
  );
}

export function useRecentTraces(options?: SWRConfiguration) {
  return useSWR<{ traces: RecentTrace[] }>(
    'recent-traces',
    () => fetcher(() => api.getRecentTraces()),
    { ...defaultOptions, ...options }
  );
}

// Traces Hooks
export function useTraces(
  params?: { limit?: number; offset?: number; server?: string; status?: string },
  options?: SWRConfiguration
) {
  const key = params ? ['traces', JSON.stringify(params)] : 'traces';
  return useSWR<TracesResponse>(
    key,
    () => fetcher(() => api.listTraces(params)),
    { ...defaultOptions, ...options }
  );
}

export function useTrace(traceId: string | null, options?: SWRConfiguration) {
  return useSWR<TraceDetail>(
    traceId ? ['trace', traceId] : null,
    () => fetcher(() => api.getTrace(traceId!)),
    { ...defaultOptions, ...options }
  );
}

export function useTraceStats(options?: SWRConfiguration) {
  return useSWR<TraceStats>(
    'trace-stats',
    () => fetcher(() => api.getTraceStats()),
    { ...defaultOptions, ...options }
  );
}

// Costs Hooks
export function useCostSummary(period?: string, options?: SWRConfiguration) {
  return useSWR<CostSummaryResponse>(
    period ? ['cost-summary', period] : 'cost-summary',
    () => fetcher(() => api.getCostSummary(period)),
    { ...defaultOptions, ...options }
  );
}

export function useCostsByServer(options?: SWRConfiguration) {
  return useSWR<CostsByServerResponse>(
    'costs-by-server',
    () => fetcher(() => api.getCostsByServer()),
    { ...defaultOptions, ...options }
  );
}

export function useCostsByTeam(options?: SWRConfiguration) {
  return useSWR<CostsByTeamResponse>(
    'costs-by-team',
    () => fetcher(() => api.getCostsByTeam()),
    { ...defaultOptions, ...options }
  );
}

export function useDailyCosts(options?: SWRConfiguration) {
  return useSWR<{ daily: DailyCost[] }>(
    'daily-costs',
    () => fetcher(() => api.getDailyCosts()),
    { ...defaultOptions, ...options }
  );
}

// API Keys Hooks
export function useApiKeys(options?: SWRConfiguration) {
  return useSWR<ApiKeysResponse>(
    'api-keys',
    () => fetcher(() => api.listApiKeys()),
    { ...defaultOptions, ...options }
  );
}

export function useApiKey(keyId: string | null, options?: SWRConfiguration) {
  return useSWR<ApiKeyDetail>(
    keyId ? ['api-key', keyId] : null,
    () => fetcher(() => api.getApiKey(keyId!)),
    { ...defaultOptions, ...options }
  );
}

// API Key Mutations
export function useCreateApiKey() {
  return useSWRMutation<ApiKeyCreated, Error, string, CreateApiKeyRequest>(
    'api-keys',
    (_, { arg }) => api.createApiKey(arg)
  );
}

export function useDeleteApiKey() {
  return useSWRMutation<{ message: string }, Error, string, string>(
    'api-keys',
    (_, { arg: keyId }) => api.deleteApiKey(keyId)
  );
}

export function useRotateApiKey() {
  return useSWRMutation<ApiKeyCreated, Error, string, string>(
    'api-keys',
    (_, { arg: keyId }) => api.rotateApiKey(keyId)
  );
}

// Alert Rules Hooks
export function useAlertRules(options?: SWRConfiguration) {
  return useSWR<AlertRulesResponse>(
    'alert-rules',
    () => fetcher(() => api.listAlertRules()),
    { ...defaultOptions, ...options }
  );
}

export function useAlertRule(ruleId: string | null, options?: SWRConfiguration) {
  return useSWR<AlertRule>(
    ruleId ? ['alert-rule', ruleId] : null,
    () => fetcher(() => api.getAlertRule(ruleId!)),
    { ...defaultOptions, ...options }
  );
}

export function useCreateAlertRule() {
  return useSWRMutation<AlertRule, Error, string, CreateAlertRuleRequest>(
    'alert-rules',
    (_, { arg }) => api.createAlertRule(arg)
  );
}

export function useUpdateAlertRule(ruleId: string) {
  return useSWRMutation<AlertRule, Error, string, UpdateAlertRuleRequest>(
    'alert-rules',
    (_, { arg }) => api.updateAlertRule(ruleId, arg)
  );
}

export function useDeleteAlertRule() {
  return useSWRMutation<{ status: string }, Error, string, string>(
    'alert-rules',
    (_, { arg: ruleId }) => api.deleteAlertRule(ruleId)
  );
}

// Alert Channels Hooks
export function useAlertChannels(options?: SWRConfiguration) {
  return useSWR<AlertChannelsResponse>(
    'alert-channels',
    () => fetcher(() => api.listAlertChannels()),
    { ...defaultOptions, ...options }
  );
}

export function useAlertChannel(channelId: string | null, options?: SWRConfiguration) {
  return useSWR<AlertChannel>(
    channelId ? ['alert-channel', channelId] : null,
    () => fetcher(() => api.getAlertChannel(channelId!)),
    { ...defaultOptions, ...options }
  );
}

export function useCreateAlertChannel() {
  return useSWRMutation<AlertChannel, Error, string, CreateAlertChannelRequest>(
    'alert-channels',
    (_, { arg }) => api.createAlertChannel(arg)
  );
}

export function useUpdateAlertChannel(channelId: string) {
  return useSWRMutation<AlertChannel, Error, string, UpdateAlertChannelRequest>(
    'alert-channels',
    (_, { arg }) => api.updateAlertChannel(channelId, arg)
  );
}

export function useDeleteAlertChannel() {
  return useSWRMutation<{ status: string }, Error, string, string>(
    'alert-channels',
    (_, { arg: channelId }) => api.deleteAlertChannel(channelId)
  );
}

export function useTestAlertChannel() {
  return useSWRMutation<{ status: string; message: string }, Error, string, string>(
    'alert-channels',
    (_, { arg: channelId }) => api.testAlertChannel(channelId)
  );
}

// Alerts Hooks
export function useAlerts(params?: AlertsFilterParams, options?: SWRConfiguration) {
  const key = params ? ['alerts', JSON.stringify(params)] : 'alerts';
  return useSWR<AlertsResponse>(
    key,
    () => fetcher(() => api.listAlerts(params)),
    { ...defaultOptions, ...options }
  );
}

export function useActiveAlerts(options?: SWRConfiguration) {
  return useSWR<ActiveAlertsResponse>(
    'active-alerts',
    () => fetcher(() => api.getActiveAlerts()),
    { ...defaultOptions, refreshInterval: 30000, ...options } // Auto-refresh every 30s
  );
}

export function useAcknowledgeAlert() {
  return useSWRMutation<Alert, Error, string, string>(
    'active-alerts',
    (_, { arg: alertId }) => api.acknowledgeAlert(alertId)
  );
}

export function useResolveAlert() {
  return useSWRMutation<Alert, Error, string, string>(
    'active-alerts',
    (_, { arg: alertId }) => api.resolveAlert(alertId)
  );
}

export function useTriggerTestAlert() {
  return useSWRMutation<Alert, Error, string, { metric?: string; value?: number } | undefined>(
    'active-alerts',
    (_, { arg }) => api.triggerTestAlert(arg)
  );
}

// Safety Policies Hooks
export function useSafetyPolicies(options?: SWRConfiguration) {
  return useSWR<SafetyPoliciesResponse>(
    'safety-policies',
    () => fetcher(() => api.listSafetyPolicies()),
    { ...defaultOptions, ...options }
  );
}

export function useSafetyPolicy(policyId: string | null, options?: SWRConfiguration) {
  return useSWR<SafetyPolicy>(
    policyId ? ['safety-policy', policyId] : null,
    () => fetcher(() => api.getSafetyPolicy(policyId!)),
    { ...defaultOptions, ...options }
  );
}

export function useCreateSafetyPolicy() {
  return useSWRMutation<SafetyPolicy, Error, string, CreateSafetyPolicyRequest>(
    'safety-policies',
    (_, { arg }) => api.createSafetyPolicy(arg)
  );
}

export function useUpdateSafetyPolicy(policyId: string) {
  return useSWRMutation<SafetyPolicy, Error, string, UpdateSafetyPolicyRequest>(
    'safety-policies',
    (_, { arg }) => api.updateSafetyPolicy(policyId, arg)
  );
}

export function useDeleteSafetyPolicy() {
  return useSWRMutation<void, Error, string, string>(
    'safety-policies',
    (_, { arg: policyId }) => api.deleteSafetyPolicy(policyId)
  );
}

// Safety Detections Hooks
export function useDetections(params?: DetectionsFilterParams, options?: SWRConfiguration) {
  const key = params ? ['detections', JSON.stringify(params)] : 'detections';
  return useSWR<DetectionsResponse>(
    key,
    () => fetcher(() => api.listDetections(params)),
    { ...defaultOptions, ...options }
  );
}

export function useSafetySummary(options?: SWRConfiguration) {
  return useSWR<SafetySummary>(
    'safety-summary',
    () => fetcher(() => api.getSafetySummary()),
    { ...defaultOptions, ...options }
  );
}

export function useTestSafetyInput() {
  return useSWRMutation<SafetyTestResponse, Error, string, SafetyTestRequest>(
    'safety-test',
    (_, { arg }) => api.testSafetyInput(arg)
  );
}

// Users Hooks
export function useUsers(params?: { limit?: number; offset?: number }, options?: SWRConfiguration) {
  const key = params ? ['users', JSON.stringify(params)] : 'users';
  return useSWR<UsersResponse>(
    key,
    () => fetcher(() => api.listUsers(params)),
    { ...defaultOptions, ...options }
  );
}

export function useUser(userId: string | null, options?: SWRConfiguration) {
  return useSWR<User>(
    userId ? ['user', userId] : null,
    () => fetcher(() => api.getUser(userId!)),
    { ...defaultOptions, ...options }
  );
}

// Invites Hooks
export function useInvites(options?: SWRConfiguration) {
  return useSWR<InvitesResponse>(
    'invites',
    () => fetcher(() => api.listInvites()),
    { ...defaultOptions, ...options }
  );
}

export function useCreateInvite() {
  return useSWRMutation<Invite, Error, string, CreateInviteRequest>(
    'invites',
    (_, { arg }) => api.createInvite(arg)
  );
}

export function useCancelInvite() {
  return useSWRMutation<{ status: string }, Error, string, string>(
    'invites',
    (_, { arg: inviteId }) => api.cancelInvite(inviteId)
  );
}

export function useResendInvite() {
  return useSWRMutation<{ status: string; invite: Invite }, Error, string, string>(
    'invites',
    (_, { arg: inviteId }) => api.resendInvite(inviteId)
  );
}

// Roles Hooks
export function useRoles(options?: SWRConfiguration) {
  return useSWR<RolesResponse>(
    'roles',
    () => fetcher(() => api.listRoles()),
    { ...defaultOptions, ...options }
  );
}

// Settings Hooks
export function useSettings(options?: SWRConfiguration) {
  return useSWR<OrgSettings>(
    'settings',
    () => fetcher(() => api.getSettings()),
    { ...defaultOptions, ...options }
  );
}

export function useUpdateSettings() {
  return useSWRMutation<OrgSettings, Error, string, UpdateSettingsRequest>(
    'settings',
    (_, { arg }) => api.updateSettings(arg)
  );
}

// SSO Provider Hooks
export function useSSOProviders(includeDisabled?: boolean, options?: SWRConfiguration) {
  const key = includeDisabled ? 'sso-providers-all' : 'sso-providers';
  return useSWR<SSOProvidersResponse>(
    key,
    () => fetcher(() => api.listSSOProviders(includeDisabled)),
    { ...defaultOptions, ...options }
  );
}

export function useSSOProvider(providerId: string | null, options?: SWRConfiguration) {
  return useSWR<SSOProvider>(
    providerId ? ['sso-provider', providerId] : null,
    () => fetcher(() => api.getSSOProvider(providerId!)),
    { ...defaultOptions, ...options }
  );
}

export function useCreateSSOProvider() {
  return useSWRMutation<SSOProvider, Error, string, CreateSSOProviderRequest>(
    'sso-providers',
    (_, { arg }) => api.createSSOProvider(arg)
  );
}

export function useUpdateSSOProvider(providerId: string) {
  return useSWRMutation<SSOProvider, Error, string, UpdateSSOProviderRequest>(
    'sso-providers',
    (_, { arg }) => api.updateSSOProvider(providerId, arg)
  );
}

export function useDeleteSSOProvider() {
  return useSWRMutation<{ status: string }, Error, string, string>(
    'sso-providers',
    (_, { arg: providerId }) => api.deleteSSOProvider(providerId)
  );
}

export function useTestSSOProvider() {
  return useSWRMutation<{ status: string; message: string }, Error, string, string>(
    'sso-providers',
    (_, { arg: providerId }) => api.testSSOProvider(providerId)
  );
}

export function useSupportedSSOProviders(options?: SWRConfiguration) {
  return useSWR<SupportedSSOProvidersResponse>(
    'supported-sso-providers',
    () => fetcher(() => api.getSupportedSSOProviders()),
    { ...defaultOptions, ...options }
  );
}

// Telemetry Config Hooks
export function useTelemetryConfigs(options?: SWRConfiguration) {
  return useSWR<TelemetryConfigsResponse>(
    'telemetry-configs',
    () => fetcher(() => api.listTelemetryConfigs()),
    { ...defaultOptions, ...options }
  );
}

export function useTelemetryConfig(configId: string | null, options?: SWRConfiguration) {
  return useSWR<TelemetryConfig>(
    configId ? ['telemetry-config', configId] : null,
    () => fetcher(() => api.getTelemetryConfig(configId!)),
    { ...defaultOptions, ...options }
  );
}

export function useCreateTelemetryConfig() {
  return useSWRMutation<TelemetryConfig, Error, string, CreateTelemetryConfigRequest>(
    'telemetry-configs',
    (_, { arg }) => api.createTelemetryConfig(arg)
  );
}

export function useUpdateTelemetryConfig(configId: string) {
  return useSWRMutation<TelemetryConfig, Error, string, UpdateTelemetryConfigRequest>(
    'telemetry-configs',
    (_, { arg }) => api.updateTelemetryConfig(configId, arg)
  );
}

export function useDeleteTelemetryConfig() {
  return useSWRMutation<{ status: string }, Error, string, string>(
    'telemetry-configs',
    (_, { arg: configId }) => api.deleteTelemetryConfig(configId)
  );
}

export function useTestTelemetryConfig() {
  return useSWRMutation<TelemetryTestResult, Error, string, string>(
    'telemetry-configs',
    (_, { arg: configId }) => api.testTelemetryConfig(configId)
  );
}

export function useSupportedExporters(options?: SWRConfiguration) {
  return useSWR<SupportedExportersResponse>(
    'supported-exporters',
    () => fetcher(() => api.getSupportedExporters()),
    { ...defaultOptions, ...options }
  );
}
