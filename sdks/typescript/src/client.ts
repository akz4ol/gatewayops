/**
 * GatewayOps SDK client
 */

import { fetch, type RequestInit, type Response } from 'undici';
import { SDK_VERSION } from './version';
import { AgentsClient } from './agents';
import type {
  GatewayOpsOptions,
  ToolDefinition,
  ToolCallResult,
  Resource,
  ResourceContent,
  Prompt,
  PromptMessage,
  Trace,
  TracePage,
  TraceFilter,
  CostSummary,
  CostPeriod,
  CostGroupBy,
} from './types';
import {
  GatewayOpsError,
  AuthenticationError,
  RateLimitError,
  NotFoundError,
  ValidationError,
  InjectionDetectedError,
  ToolAccessDeniedError,
  ServerError,
  NetworkError,
  TimeoutError,
} from './errors';

const DEFAULT_BASE_URL = 'https://api.gatewayops.com';
const DEFAULT_TIMEOUT = 30000;
const DEFAULT_MAX_RETRIES = 3;

/**
 * Main GatewayOps client
 *
 * @example
 * ```typescript
 * const gw = new GatewayOps({ apiKey: 'gwo_prd_...' });
 * const result = await gw.mcp('filesystem').tools.call('read_file', { path: '/data.csv' });
 * ```
 */
export class GatewayOps {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly timeout: number;
  private readonly maxRetries: number;
  private traceContext?: string;

  constructor(options: GatewayOpsOptions) {
    this.apiKey = options.apiKey;
    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/$/, '');
    this.timeout = options.timeout ?? DEFAULT_TIMEOUT;
    this.maxRetries = options.maxRetries ?? DEFAULT_MAX_RETRIES;
  }

  /**
   * Get an MCP client for a specific server
   */
  mcp(server: string): MCPClient {
    return new MCPClient(this, server);
  }

  /**
   * Create a trace context for correlating operations
   */
  trace(name: string): TraceContext {
    const traceId = crypto.randomUUID();
    return new TraceContext(this, traceId, name);
  }

  /**
   * Get the traces client
   */
  get traces(): TracesClient {
    return new TracesClient(this);
  }

  /**
   * Get the costs client
   */
  get costs(): CostsClient {
    return new CostsClient(this);
  }

  /**
   * Get the agents client for agent platform integration
   */
  get agents(): AgentsClient {
    return new AgentsClient(this);
  }

  /**
   * Set the trace context for subsequent requests
   * @internal
   */
  setTraceContext(traceId: string | undefined): void {
    this.traceContext = traceId;
  }

  /**
   * Make an HTTP request to the API
   * @internal
   */
  async request<T>(
    method: string,
    path: string,
    options?: { data?: Record<string, unknown>; params?: Record<string, unknown> }
  ): Promise<T> {
    const url = new URL(path, this.baseUrl);

    if (options?.params) {
      for (const [key, value] of Object.entries(options.params)) {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      }
    }

    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json',
      'User-Agent': `gatewayops-typescript/${SDK_VERSION}`,
    };

    if (this.traceContext) {
      headers['X-Trace-ID'] = this.traceContext;
    }

    const requestInit: RequestInit = {
      method,
      headers,
      signal: AbortSignal.timeout(this.timeout),
    };

    if (options?.data) {
      requestInit.body = JSON.stringify(options.data);
    }

    let lastError: Error | undefined;
    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        const response = await fetch(url.toString(), requestInit);
        return await this.handleResponse<T>(response);
      } catch (error) {
        lastError = error as Error;

        // Don't retry on client errors
        if (error instanceof GatewayOpsError && error.statusCode && error.statusCode < 500) {
          throw error;
        }

        // Don't retry on timeout
        if (error instanceof TimeoutError) {
          throw error;
        }

        // Wait before retrying
        if (attempt < this.maxRetries) {
          await this.sleep(Math.pow(2, attempt) * 1000);
        }
      }
    }

    throw lastError ?? new NetworkError('Request failed');
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    let data: Record<string, unknown> = {};

    try {
      data = (await response.json()) as Record<string, unknown>;
    } catch {
      // Response may not be JSON
    }

    if (response.status >= 400) {
      this.raiseForError(response.status, data);
    }

    return data as T;
  }

  private raiseForError(statusCode: number, data: Record<string, unknown>): never {
    const error = data.error as Record<string, unknown> | undefined;
    const errorCode = (error?.code as string) ?? 'unknown';
    const message = (error?.message as string) ?? 'Unknown error';
    const details = (error?.details as Record<string, unknown>) ?? {};

    switch (statusCode) {
      case 401:
        throw new AuthenticationError(message, details);

      case 403:
        if (errorCode === 'tool_access_denied') {
          throw new ToolAccessDeniedError(message, {
            mcpServer: details.mcp_server as string | undefined,
            toolName: details.tool_name as string | undefined,
            requiresApproval: (details.requires_approval as boolean) ?? false,
            details,
          });
        }
        throw new GatewayOpsError(message, { code: errorCode, statusCode: 403, details });

      case 404:
        throw new NotFoundError(message, { details });

      case 429:
        throw new RateLimitError(message, {
          retryAfter: details['Retry-After'] as number | undefined,
          details,
        });

      case 400:
        if (errorCode === 'injection_detected') {
          throw new InjectionDetectedError(message, {
            pattern: details.pattern as string | undefined,
            severity: details.severity as string | undefined,
            details,
          });
        }
        throw new ValidationError(message, { details });

      default:
        if (statusCode >= 500) {
          throw new ServerError(message, { code: errorCode, details });
        }
        throw new GatewayOpsError(message, { code: errorCode, statusCode, details });
    }
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * Client for MCP operations on a specific server
 */
export class MCPClient {
  constructor(
    private readonly client: GatewayOps,
    private readonly server: string
  ) {}

  /**
   * Get the tools client
   */
  get tools(): ToolsClient {
    return new ToolsClient(this.client, this.server);
  }

  /**
   * Get the resources client
   */
  get resources(): ResourcesClient {
    return new ResourcesClient(this.client, this.server);
  }

  /**
   * Get the prompts client
   */
  get prompts(): PromptsClient {
    return new PromptsClient(this.client, this.server);
  }
}

/**
 * Client for MCP tool operations
 */
export class ToolsClient {
  constructor(
    private readonly client: GatewayOps,
    private readonly server: string
  ) {}

  /**
   * List available tools
   */
  async list(): Promise<ToolDefinition[]> {
    const response = await this.client.request<{ tools: ToolDefinition[] }>(
      'POST',
      `/v1/mcp/${this.server}/tools/list`
    );
    return response.tools ?? [];
  }

  /**
   * Call a tool
   */
  async call(tool: string, args: Record<string, unknown> = {}): Promise<ToolCallResult> {
    return this.client.request<ToolCallResult>('POST', `/v1/mcp/${this.server}/tools/call`, {
      data: { tool, arguments: args },
    });
  }
}

/**
 * Client for MCP resource operations
 */
export class ResourcesClient {
  constructor(
    private readonly client: GatewayOps,
    private readonly server: string
  ) {}

  /**
   * List available resources
   */
  async list(): Promise<Resource[]> {
    const response = await this.client.request<{ resources: Resource[] }>(
      'POST',
      `/v1/mcp/${this.server}/resources/list`
    );
    return response.resources ?? [];
  }

  /**
   * Read a resource
   */
  async read(uri: string): Promise<ResourceContent> {
    return this.client.request<ResourceContent>('POST', `/v1/mcp/${this.server}/resources/read`, {
      data: { uri },
    });
  }
}

/**
 * Client for MCP prompt operations
 */
export class PromptsClient {
  constructor(
    private readonly client: GatewayOps,
    private readonly server: string
  ) {}

  /**
   * List available prompts
   */
  async list(): Promise<Prompt[]> {
    const response = await this.client.request<{ prompts: Prompt[] }>(
      'POST',
      `/v1/mcp/${this.server}/prompts/list`
    );
    return response.prompts ?? [];
  }

  /**
   * Get a prompt with arguments
   */
  async get(name: string, args: Record<string, unknown> = {}): Promise<PromptMessage[]> {
    const response = await this.client.request<{ messages: PromptMessage[] }>(
      'POST',
      `/v1/mcp/${this.server}/prompts/get`,
      { data: { name, arguments: args } }
    );
    return response.messages ?? [];
  }
}

/**
 * Client for trace operations
 */
export class TracesClient {
  constructor(private readonly client: GatewayOps) {}

  /**
   * List traces with optional filters
   */
  async list(filter?: TraceFilter): Promise<TracePage> {
    const params: Record<string, unknown> = {
      limit: filter?.limit ?? 50,
      offset: filter?.offset ?? 0,
    };

    if (filter?.mcpServer) params.mcp_server = filter.mcpServer;
    if (filter?.operation) params.operation = filter.operation;
    if (filter?.status) params.status = filter.status;
    if (filter?.startTime) params.start_time = filter.startTime.toISOString();
    if (filter?.endTime) params.end_time = filter.endTime.toISOString();

    return this.client.request<TracePage>('GET', '/v1/traces', { params });
  }

  /**
   * Get a specific trace by ID
   */
  async get(traceId: string): Promise<Trace> {
    return this.client.request<Trace>('GET', `/v1/traces/${traceId}`);
  }
}

/**
 * Client for cost operations
 */
export class CostsClient {
  constructor(private readonly client: GatewayOps) {}

  /**
   * Get cost summary for a period
   */
  async summary(options?: { period?: CostPeriod; groupBy?: CostGroupBy }): Promise<CostSummary> {
    const params: Record<string, unknown> = {
      period: options?.period ?? 'month',
    };

    if (options?.groupBy) {
      params.group_by = options.groupBy;
    }

    return this.client.request<CostSummary>('GET', '/v1/costs/summary', { params });
  }

  /**
   * Get costs grouped by MCP server
   */
  async byServer(period: CostPeriod = 'month'): Promise<CostSummary> {
    return this.summary({ period, groupBy: 'server' });
  }

  /**
   * Get costs grouped by team
   */
  async byTeam(period: CostPeriod = 'month'): Promise<CostSummary> {
    return this.summary({ period, groupBy: 'team' });
  }

  /**
   * Get costs grouped by tool
   */
  async byTool(period: CostPeriod = 'month'): Promise<CostSummary> {
    return this.summary({ period, groupBy: 'tool' });
  }
}

/**
 * Context for tracing operations
 */
export class TraceContext {
  constructor(
    private readonly client: GatewayOps,
    public readonly traceId: string,
    public readonly name: string
  ) {}

  /**
   * Execute a function within this trace context
   */
  async run<T>(fn: () => Promise<T>): Promise<T> {
    this.client.setTraceContext(this.traceId);
    try {
      return await fn();
    } finally {
      this.client.setTraceContext(undefined);
    }
  }
}
