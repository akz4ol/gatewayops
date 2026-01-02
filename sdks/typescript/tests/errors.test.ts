/**
 * Tests for GatewayOps SDK errors
 */

import { describe, it, expect } from 'vitest';
import {
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
} from '../src/errors';

describe('GatewayOpsError', () => {
  it('should store message', () => {
    const error = new GatewayOpsError('Something went wrong');
    expect(error.message).toBe('Something went wrong');
  });

  it('should store code', () => {
    const error = new GatewayOpsError('Failed', { code: 'some_error' });
    expect(error.code).toBe('some_error');
  });

  it('should store status code', () => {
    const error = new GatewayOpsError('Failed', { statusCode: 500 });
    expect(error.statusCode).toBe(500);
  });

  it('should store details', () => {
    const error = new GatewayOpsError('Failed', { details: { key: 'value' } });
    expect(error.details).toEqual({ key: 'value' });
  });

  it('should be instanceof Error', () => {
    const error = new GatewayOpsError('Test');
    expect(error).toBeInstanceOf(Error);
  });
});

describe('AuthenticationError', () => {
  it('should have authentication message', () => {
    const error = new AuthenticationError('Invalid API key');
    expect(error.message).toBe('Invalid API key');
  });

  it('should have 401 status code', () => {
    const error = new AuthenticationError('Test');
    expect(error.statusCode).toBe(401);
  });

  it('should have unauthorized code', () => {
    const error = new AuthenticationError('Test');
    expect(error.code).toBe('unauthorized');
  });
});

describe('RateLimitError', () => {
  it('should have rate limit message', () => {
    const error = new RateLimitError('Too many requests');
    expect(error.message).toBe('Too many requests');
  });

  it('should have 429 status code', () => {
    const error = new RateLimitError('Test');
    expect(error.statusCode).toBe(429);
  });

  it('should store retry after', () => {
    const error = new RateLimitError('Test', { retryAfter: 60 });
    expect(error.retryAfter).toBe(60);
  });
});

describe('NotFoundError', () => {
  it('should have not found message', () => {
    const error = new NotFoundError('Resource not found');
    expect(error.message).toBe('Resource not found');
  });

  it('should have 404 status code', () => {
    const error = new NotFoundError('Test');
    expect(error.statusCode).toBe(404);
  });
});

describe('ValidationError', () => {
  it('should have validation message', () => {
    const error = new ValidationError('Invalid input');
    expect(error.message).toBe('Invalid input');
  });

  it('should have 400 status code', () => {
    const error = new ValidationError('Test');
    expect(error.statusCode).toBe(400);
  });
});

describe('InjectionDetectedError', () => {
  it('should have injection message', () => {
    const error = new InjectionDetectedError('Prompt injection detected');
    expect(error.message).toBe('Prompt injection detected');
  });

  it('should have 400 status code', () => {
    const error = new InjectionDetectedError('Test');
    expect(error.statusCode).toBe(400);
  });

  it('should store pattern', () => {
    const error = new InjectionDetectedError('Test', { pattern: 'ignore instructions' });
    expect(error.pattern).toBe('ignore instructions');
  });

  it('should store severity', () => {
    const error = new InjectionDetectedError('Test', { severity: 'high' });
    expect(error.severity).toBe('high');
  });
});

describe('ToolAccessDeniedError', () => {
  it('should have access denied message', () => {
    const error = new ToolAccessDeniedError('Tool access denied');
    expect(error.message).toBe('Tool access denied');
  });

  it('should have 403 status code', () => {
    const error = new ToolAccessDeniedError('Test');
    expect(error.statusCode).toBe(403);
  });

  it('should store mcp server', () => {
    const error = new ToolAccessDeniedError('Test', { mcpServer: 'filesystem' });
    expect(error.mcpServer).toBe('filesystem');
  });

  it('should store tool name', () => {
    const error = new ToolAccessDeniedError('Test', { toolName: 'delete_file' });
    expect(error.toolName).toBe('delete_file');
  });

  it('should store requires approval', () => {
    const error = new ToolAccessDeniedError('Test', { requiresApproval: true });
    expect(error.requiresApproval).toBe(true);
  });
});

describe('ServerError', () => {
  it('should have server error message', () => {
    const error = new ServerError('Internal server error');
    expect(error.message).toBe('Internal server error');
  });

  it('should have 500 status code', () => {
    const error = new ServerError('Test');
    expect(error.statusCode).toBe(500);
  });

  it('should accept custom code', () => {
    const error = new ServerError('Test', { code: 'database_error' });
    expect(error.code).toBe('database_error');
  });
});

describe('TimeoutError', () => {
  it('should have timeout message', () => {
    const error = new TimeoutError('Request timed out');
    expect(error.message).toBe('Request timed out');
  });

  it('should have 408 status code', () => {
    const error = new TimeoutError('Test');
    expect(error.statusCode).toBe(408);
  });
});

describe('NetworkError', () => {
  it('should have network error message', () => {
    const error = new NetworkError('Connection failed');
    expect(error.message).toBe('Connection failed');
  });

  it('should not have status code', () => {
    const error = new NetworkError('Test');
    expect(error.statusCode).toBeUndefined();
  });
});
