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
