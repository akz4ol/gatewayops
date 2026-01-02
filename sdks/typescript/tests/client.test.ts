/**
 * Tests for GatewayOps SDK client
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GatewayOps, MCPClient, TracesClient, CostsClient, TraceContext } from '../src/client';

// Mock fetch
vi.mock('undici', () => ({
  fetch: vi.fn(),
}));

import { fetch } from 'undici';
const mockFetch = vi.mocked(fetch);

describe('GatewayOps', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe('constructor', () => {
    it('should use default base URL', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      expect((client as any).baseUrl).toBe('https://api.gatewayops.com');
    });

    it('should accept custom base URL', () => {
      const client = new GatewayOps({ apiKey: 'test', baseUrl: 'https://custom.api.com/' });
      expect((client as any).baseUrl).toBe('https://custom.api.com');
    });

    it('should use default timeout', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      expect((client as any).timeout).toBe(30000);
    });

    it('should accept custom timeout', () => {
      const client = new GatewayOps({ apiKey: 'test', timeout: 60000 });
      expect((client as any).timeout).toBe(60000);
    });

    it('should use default max retries', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      expect((client as any).maxRetries).toBe(3);
    });

    it('should accept custom max retries', () => {
      const client = new GatewayOps({ apiKey: 'test', maxRetries: 5 });
      expect((client as any).maxRetries).toBe(5);
    });
  });

  describe('mcp()', () => {
    it('should return MCPClient instance', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      const mcp = client.mcp('filesystem');
      expect(mcp).toBeInstanceOf(MCPClient);
    });

    it('should create MCPClient with correct server', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      const mcp = client.mcp('database');
      expect((mcp as any).server).toBe('database');
    });
  });

  describe('traces', () => {
    it('should return TracesClient instance', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      expect(client.traces).toBeInstanceOf(TracesClient);
    });
  });

  describe('costs', () => {
    it('should return CostsClient instance', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      expect(client.costs).toBeInstanceOf(CostsClient);
    });
  });

  describe('trace()', () => {
    it('should return TraceContext', () => {
      const client = new GatewayOps({ apiKey: 'test' });
      const ctx = client.trace('test-operation');
      expect(ctx).toBeInstanceOf(TraceContext);
      expect(ctx.name).toBe('test-operation');
      expect(ctx.traceId).toBeDefined();
    });
  });
});

describe('MCPClient', () => {
  let client: GatewayOps;

  beforeEach(() => {
    client = new GatewayOps({ apiKey: 'test' });
    vi.resetAllMocks();
  });

  describe('tools', () => {
    it('should have tools client', () => {
      const mcp = client.mcp('filesystem');
      expect(mcp.tools).toBeDefined();
    });

    it('should call tools.list with correct endpoint', async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        json: async () => ({ tools: [] }),
      } as any);

      const mcp = client.mcp('filesystem');
      await mcp.tools.list();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/v1/mcp/filesystem/tools/list'),
        expect.any(Object)
      );
    });

    it('should call tools.call with correct endpoint', async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        json: async () => ({ content: 'result', isError: false }),
      } as any);

      const mcp = client.mcp('filesystem');
      await mcp.tools.call('read_file', { path: '/test.txt' });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/v1/mcp/filesystem/tools/call'),
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('read_file'),
        })
      );
    });
  });

  describe('resources', () => {
    it('should have resources client', () => {
      const mcp = client.mcp('filesystem');
      expect(mcp.resources).toBeDefined();
    });
  });

  describe('prompts', () => {
    it('should have prompts client', () => {
      const mcp = client.mcp('filesystem');
      expect(mcp.prompts).toBeDefined();
    });
  });
});

describe('TracesClient', () => {
  let client: GatewayOps;

  beforeEach(() => {
    client = new GatewayOps({ apiKey: 'test' });
    vi.resetAllMocks();
  });

  it('should call list with correct endpoint', async () => {
    mockFetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({ traces: [], total: 0, limit: 50, offset: 0 }),
    } as any);

    await client.traces.list({ limit: 10 });

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/traces'),
      expect.any(Object)
    );
  });

  it('should call get with correct endpoint', async () => {
    mockFetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({
        id: 'tr_123',
        orgId: 'org_1',
        mcpServer: 'filesystem',
        operation: 'tools/call',
        status: 'success',
        startTime: '2026-01-01T00:00:00Z',
      }),
    } as any);

    await client.traces.get('tr_123');

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/traces/tr_123'),
      expect.any(Object)
    );
  });
});

describe('CostsClient', () => {
  let client: GatewayOps;

  beforeEach(() => {
    client = new GatewayOps({ apiKey: 'test' });
    vi.resetAllMocks();
  });

  it('should call summary with correct endpoint', async () => {
    mockFetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({
        total_cost: 100,
        total_requests: 1000,
        avg_cost_per_request: 0.1,
        period: 'month',
        start_date: '2025-12-01T00:00:00Z',
        end_date: '2026-01-01T00:00:00Z',
      }),
    } as any);

    await client.costs.summary({ period: 'month' });

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/costs/summary'),
      expect.any(Object)
    );
  });

  it('should call byServer with correct params', async () => {
    mockFetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({
        total_cost: 100,
        total_requests: 1000,
        avg_cost_per_request: 0.1,
        period: 'month',
        start_date: '2025-12-01T00:00:00Z',
        end_date: '2026-01-01T00:00:00Z',
      }),
    } as any);

    await client.costs.byServer('week');

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('period=week'),
      expect.any(Object)
    );
  });
});

describe('TraceContext', () => {
  it('should run function with trace context', async () => {
    const client = new GatewayOps({ apiKey: 'test' });
    const ctx = client.trace('test-operation');

    let capturedTraceId: string | undefined;

    await ctx.run(async () => {
      capturedTraceId = (client as any).traceContext;
      return 'result';
    });

    expect(capturedTraceId).toBe(ctx.traceId);
  });

  it('should clear trace context after run', async () => {
    const client = new GatewayOps({ apiKey: 'test' });
    const ctx = client.trace('test-operation');

    await ctx.run(async () => 'result');

    expect((client as any).traceContext).toBeUndefined();
  });
});
