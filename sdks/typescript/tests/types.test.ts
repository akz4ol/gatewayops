/**
 * Tests for GatewayOps SDK types
 */

import { describe, it, expect } from 'vitest';
import type { TracePage, CostSummary, ToolDefinition, ToolCallResult, Resource, Span } from '../src/types';
import { hasMoreTraces, getTotalCost, getRequestCount } from '../src/types/traces';
import { getTotalCost as getCostTotal, getRequestCount as getCostRequestCount } from '../src/types/costs';

describe('TracePage', () => {
  it('should handle null traces array', () => {
    const page: TracePage = {
      traces: null,
      total: 0,
      limit: 20,
      offset: 0,
    };
    expect(page.traces).toBeNull();
    expect(page.total).toBe(0);
  });

  it('should handle empty traces array', () => {
    const page: TracePage = {
      traces: [],
      total: 0,
      limit: 20,
      offset: 0,
    };
    expect(page.traces).toEqual([]);
  });
});

describe('hasMoreTraces', () => {
  it('should return false when no more traces', () => {
    const page: TracePage = {
      traces: [],
      total: 0,
      limit: 20,
      offset: 0,
    };
    expect(hasMoreTraces(page)).toBe(false);
  });

  it('should return true when more traces exist', () => {
    const page: TracePage = {
      traces: [],
      total: 100,
      limit: 20,
      offset: 0,
    };
    expect(hasMoreTraces(page)).toBe(true);
  });

  it('should return false at end of results', () => {
    const page: TracePage = {
      traces: [],
      total: 100,
      limit: 20,
      offset: 80,
    };
    expect(hasMoreTraces(page)).toBe(false);
  });

  it('should return false when exactly at total', () => {
    const page: TracePage = {
      traces: [],
      total: 20,
      limit: 20,
      offset: 0,
    };
    expect(hasMoreTraces(page)).toBe(false);
  });
});

describe('CostSummary', () => {
  it('should use snake_case field names from API', () => {
    const summary: CostSummary = {
      total_cost: 123.45,
      total_requests: 1000,
      avg_cost_per_request: 0.12345,
      period: 'month',
      start_date: '2025-12-01T00:00:00Z',
      end_date: '2026-01-01T00:00:00Z',
    };
    expect(summary.total_cost).toBe(123.45);
    expect(summary.total_requests).toBe(1000);
    expect(summary.avg_cost_per_request).toBe(0.12345);
    expect(summary.period).toBe('month');
  });
});

describe('getTotalCost helper', () => {
  it('should return total_cost from summary', () => {
    const summary: CostSummary = {
      total_cost: 99.99,
      total_requests: 500,
      avg_cost_per_request: 0.2,
      period: 'week',
      start_date: '2025-12-25T00:00:00Z',
      end_date: '2026-01-01T00:00:00Z',
    };
    expect(getCostTotal(summary)).toBe(99.99);
  });
});

describe('getRequestCount helper', () => {
  it('should return total_requests from summary', () => {
    const summary: CostSummary = {
      total_cost: 100,
      total_requests: 750,
      avg_cost_per_request: 0.133,
      period: 'day',
      start_date: '2026-01-01T00:00:00Z',
      end_date: '2026-01-02T00:00:00Z',
    };
    expect(getCostRequestCount(summary)).toBe(750);
  });
});

describe('ToolDefinition', () => {
  it('should represent basic tool', () => {
    const tool: ToolDefinition = {
      name: 'read_file',
      description: 'Read a file from the filesystem',
    };
    expect(tool.name).toBe('read_file');
    expect(tool.description).toBe('Read a file from the filesystem');
  });

  it('should represent tool with input schema', () => {
    const tool: ToolDefinition = {
      name: 'write_file',
      description: 'Write content to a file',
      inputSchema: {
        type: 'object',
        properties: {
          path: { type: 'string' },
          content: { type: 'string' },
        },
        required: ['path', 'content'],
      },
    };
    expect(tool.inputSchema).toBeDefined();
    expect(tool.inputSchema?.type).toBe('object');
  });
});

describe('ToolCallResult', () => {
  it('should represent successful result', () => {
    const result: ToolCallResult = {
      content: { data: 'file contents here' },
      isError: false,
      traceId: 'tr_abc123',
      durationMs: 45,
    };
    expect(result.content).toEqual({ data: 'file contents here' });
    expect(result.isError).toBe(false);
    expect(result.traceId).toBe('tr_abc123');
    expect(result.durationMs).toBe(45);
  });

  it('should represent error result', () => {
    const result: ToolCallResult = {
      content: 'File not found',
      isError: true,
    };
    expect(result.isError).toBe(true);
  });
});

describe('Resource', () => {
  it('should represent basic resource', () => {
    const resource: Resource = {
      uri: 'file:///data/report.csv',
      name: 'report.csv',
      description: 'Monthly report',
      mimeType: 'text/csv',
    };
    expect(resource.uri).toBe('file:///data/report.csv');
    expect(resource.name).toBe('report.csv');
    expect(resource.mimeType).toBe('text/csv');
  });
});

describe('Span', () => {
  it('should represent basic span', () => {
    const span: Span = {
      id: 'span_123',
      traceId: 'tr_abc',
      name: 'authenticate',
      kind: 'internal',
      status: 'ok',
      startTime: new Date('2026-01-01T00:00:00Z'),
      durationMs: 5,
    };
    expect(span.id).toBe('span_123');
    expect(span.traceId).toBe('tr_abc');
    expect(span.name).toBe('authenticate');
    expect(span.durationMs).toBe(5);
  });

  it('should represent span with parent', () => {
    const span: Span = {
      id: 'span_456',
      traceId: 'tr_abc',
      parentSpanId: 'span_123',
      name: 'validate_request',
      kind: 'internal',
      status: 'ok',
      startTime: new Date('2026-01-01T00:00:00Z'),
    };
    expect(span.parentSpanId).toBe('span_123');
  });
});
