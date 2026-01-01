/**
 * GatewayOps TypeScript SDK
 *
 * Official TypeScript SDK for the GatewayOps MCP Gateway.
 *
 * @example
 * ```typescript
 * import { GatewayOps } from '@gatewayops/sdk';
 *
 * const gw = new GatewayOps({ apiKey: 'gwo_prd_...' });
 * const result = await gw.mcp('filesystem').tools.call('read_file', { path: '/data.csv' });
 * ```
 *
 * @packageDocumentation
 */

export { GatewayOps, MCPClient, ToolsClient, ResourcesClient, PromptsClient, TracesClient, CostsClient, TraceContext } from './client';

export {
  GatewayOpsError,
  AuthenticationError,
  RateLimitError,
  NotFoundError,
  ValidationError,
  InjectionDetectedError,
  ToolAccessDeniedError,
  ServerError,
  TimeoutError,
  NetworkError,
} from './errors';

export type {
  // Options
  GatewayOpsOptions,
  APIError,
  APIResponse,

  // MCP types
  ToolDefinition,
  ToolCallResult,
  Resource,
  ResourceContent,
  Prompt,
  PromptArgument,
  PromptMessage,

  // Trace types
  Trace,
  TracePage,
  TraceFilter,
  TraceStatus,
  Span,
  SpanKind,
  SpanStatus,
  SpanEvent,

  // Cost types
  CostSummary,
  CostBreakdown,
  CostPeriod,
  CostGroupBy,

  // Other
  APIKey,
} from './types';
