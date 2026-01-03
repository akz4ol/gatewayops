/**
 * Agent platform support for GatewayOps SDK
 */

import { GatewayOps } from './client';

export interface ConnectOptions {
  agentId?: string;
  platform: string;
  capabilities?: string[];
  transport?: 'http' | 'websocket' | 'sse';
  callbackUrl?: string;
  metadata?: Record<string, unknown>;
}

export interface ConnectionInfo {
  connectionId: string;
  gatewayUrl: string;
  availableServers: ServerInfo[];
  rateLimits: RateLimits;
}

export interface ServerInfo {
  name: string;
  description?: string;
  tools: number;
  resources: number;
}

export interface RateLimits {
  requests_per_minute: number;
  tokens_per_minute?: number;
}

export interface ToolCall {
  id: string;
  server: string;
  tool: string;
  arguments: Record<string, unknown>;
}

export interface ExecuteOptions {
  calls: ToolCall[];
  executionMode?: 'parallel' | 'sequential';
  timeoutMs?: number;
  stream?: boolean;
}

export interface ToolResult {
  id: string;
  status: 'success' | 'error' | 'timeout';
  content?: ContentBlock[];
  error?: ErrorInfo;
  duration_ms: number;
  cost: number;
}

export interface ContentBlock {
  type: 'text' | 'image' | 'resource';
  text?: string;
  data?: string;
  uri?: string;
}

export interface ErrorInfo {
  code: string;
  message: string;
  details?: unknown;
}

export interface ExecuteResult {
  results: ToolResult[];
  traceId: string;
  totalCost: number;
}

export interface WSMessage {
  type: string;
  id?: string;
  payload?: unknown;
}

export interface ProgressEvent {
  callId: string;
  progress: number;
  message?: string;
}

export interface StreamEvent {
  event: 'start' | 'progress' | 'chunk' | 'complete' | 'error' | 'done';
  data: unknown;
}

/**
 * Agent connection for real-time communication
 */
export class AgentConnection {
  private ws: WebSocket | null = null;
  private messageHandlers: Map<string, (result: ToolResult) => void> = new Map();
  private progressHandlers: Map<string, (progress: ProgressEvent) => void> = new Map();

  constructor(
    private client: GatewayOps,
    public readonly connectionId: string,
    public readonly gatewayUrl: string,
    public readonly availableServers: ServerInfo[]
  ) {}

  /**
   * Execute tool calls via HTTP (batch)
   */
  async execute(options: ExecuteOptions): Promise<ExecuteResult> {
    const response = await this.client.request<{
      results: ToolResult[];
      trace_id: string;
      total_cost: number;
    }>('POST', '/v1/execute', {
      data: {
        connection_id: this.connectionId,
        calls: options.calls,
        execution_mode: options.executionMode || 'parallel',
        timeout_ms: options.timeoutMs || 30000,
      },
    });

    return {
      results: response.results,
      traceId: response.trace_id,
      totalCost: response.total_cost,
    };
  }

  /**
   * Execute a single tool call
   */
  async callTool(
    server: string,
    tool: string,
    args: Record<string, unknown> = {}
  ): Promise<ToolResult> {
    const callId = 'call_' + Math.random().toString(36).substr(2, 9);
    const result = await this.execute({
      calls: [{ id: callId, server, tool, arguments: args }],
    });
    return result.results[0];
  }

  /**
   * Disconnect the connection
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.messageHandlers.clear();
    this.progressHandlers.clear();
  }
}

/**
 * Agent platform client
 */
export class AgentsClient {
  constructor(private client: GatewayOps) {}

  /**
   * Connect to the gateway as an agent platform
   */
  async connect(options: ConnectOptions): Promise<AgentConnection> {
    const response = await this.client.request<{
      connection_id: string;
      gateway_url: string;
      available_servers: ServerInfo[];
      rate_limits: RateLimits;
    }>('POST', '/v1/agents/connect', {
      data: {
        agent_id: options.agentId,
        platform: options.platform,
        capabilities: options.capabilities || ['tool_calling'],
        transport: options.transport || 'http',
        callback_url: options.callbackUrl,
        metadata: options.metadata,
      },
    });

    return new AgentConnection(
      this.client,
      response.connection_id,
      response.gateway_url,
      response.available_servers
    );
  }

  /**
   * Get connection statistics
   */
  async getStats(): Promise<{
    active: number;
    total: number;
    messages: number;
    byPlatform: Record<string, number>;
  }> {
    const response = await this.client.request<{
      active: number;
      total: number;
      messages: number;
      by_platform: Record<string, number>;
    }>('GET', '/v1/agents/stats');

    return {
      active: response.active,
      total: response.total,
      messages: response.messages,
      byPlatform: response.by_platform,
    };
  }

  /**
   * List available tools in OpenAI format
   */
  async listTools(): Promise<OpenAITool[]> {
    const response = await this.client.request<{ tools: OpenAITool[] }>(
      'GET',
      '/v1/mcp/tools'
    );
    return response.tools;
  }
}

export interface OpenAITool {
  type: 'function';
  function: {
    name: string;
    description: string;
    parameters: {
      type: 'object';
      properties: Record<string, {
        type: string;
        description: string;
      }>;
      required: string[];
    };
  };
}
