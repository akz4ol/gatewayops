/**
 * Trace and span types for distributed tracing
 */

/**
 * A span within a trace
 */
export interface Span {
  id: string;
  traceId: string;
  parentSpanId?: string;
  name: string;
  kind: SpanKind;
  status: SpanStatus;
  startTime: Date;
  endTime?: Date;
  durationMs?: number;
  attributes?: Record<string, unknown>;
  events?: SpanEvent[];
}

export type SpanKind = 'internal' | 'server' | 'client' | 'producer' | 'consumer';
export type SpanStatus = 'unset' | 'ok' | 'error';

/**
 * Event within a span
 */
export interface SpanEvent {
  name: string;
  timestamp: Date;
  attributes?: Record<string, unknown>;
}

/**
 * A distributed trace
 */
export interface Trace {
  id: string;
  orgId: string;
  apiKeyId?: string;
  mcpServer: string;
  operation: string;
  status: TraceStatus;
  startTime: Date;
  endTime?: Date;
  durationMs?: number;
  spans?: Span[];
  errorMessage?: string;
  cost?: number;
}

export type TraceStatus = 'success' | 'error' | 'timeout';

/**
 * Paginated list of traces
 */
export interface TracePage {
  traces: Trace[] | null;
  total: number;
  limit: number;
  offset: number;
}

/**
 * Helper to check if there are more traces
 */
export function hasMoreTraces(page: TracePage): boolean {
  return page.offset + page.limit < page.total;
}

/**
 * Filter options for listing traces
 */
export interface TraceFilter {
  mcpServer?: string;
  operation?: string;
  status?: TraceStatus;
  startTime?: Date;
  endTime?: Date;
  limit?: number;
  offset?: number;
}
