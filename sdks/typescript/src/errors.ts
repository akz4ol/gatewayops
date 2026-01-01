/**
 * GatewayOps SDK error classes
 */

/**
 * Base error for all GatewayOps SDK errors
 */
export class GatewayOpsError extends Error {
  readonly code: string;
  readonly statusCode?: number;
  readonly details: Record<string, unknown>;

  constructor(
    message: string,
    options?: {
      code?: string;
      statusCode?: number;
      details?: Record<string, unknown>;
    }
  ) {
    super(message);
    this.name = 'GatewayOpsError';
    this.code = options?.code ?? 'unknown';
    this.statusCode = options?.statusCode;
    this.details = options?.details ?? {};
    Object.setPrototypeOf(this, GatewayOpsError.prototype);
  }
}

/**
 * Raised when authentication fails (401)
 */
export class AuthenticationError extends GatewayOpsError {
  constructor(message = 'Authentication failed', details?: Record<string, unknown>) {
    super(message, { code: 'unauthorized', statusCode: 401, details });
    this.name = 'AuthenticationError';
    Object.setPrototypeOf(this, AuthenticationError.prototype);
  }
}

/**
 * Raised when rate limit is exceeded (429)
 */
export class RateLimitError extends GatewayOpsError {
  readonly retryAfter?: number;

  constructor(
    message = 'Rate limit exceeded',
    options?: { retryAfter?: number; details?: Record<string, unknown> }
  ) {
    super(message, { code: 'rate_limit_exceeded', statusCode: 429, details: options?.details });
    this.name = 'RateLimitError';
    this.retryAfter = options?.retryAfter;
    Object.setPrototypeOf(this, RateLimitError.prototype);
  }
}

/**
 * Raised when a resource is not found (404)
 */
export class NotFoundError extends GatewayOpsError {
  readonly resourceType?: string;
  readonly resourceId?: string;

  constructor(
    message = 'Resource not found',
    options?: { resourceType?: string; resourceId?: string; details?: Record<string, unknown> }
  ) {
    super(message, { code: 'not_found', statusCode: 404, details: options?.details });
    this.name = 'NotFoundError';
    this.resourceType = options?.resourceType;
    this.resourceId = options?.resourceId;
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * Raised when request validation fails (400)
 */
export class ValidationError extends GatewayOpsError {
  readonly field?: string;

  constructor(
    message = 'Validation error',
    options?: { field?: string; details?: Record<string, unknown> }
  ) {
    super(message, { code: 'validation_error', statusCode: 400, details: options?.details });
    this.name = 'ValidationError';
    this.field = options?.field;
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

/**
 * Raised when prompt injection is detected (400)
 */
export class InjectionDetectedError extends GatewayOpsError {
  readonly pattern?: string;
  readonly severity?: string;

  constructor(
    message = 'Potential prompt injection detected',
    options?: { pattern?: string; severity?: string; details?: Record<string, unknown> }
  ) {
    super(message, { code: 'injection_detected', statusCode: 400, details: options?.details });
    this.name = 'InjectionDetectedError';
    this.pattern = options?.pattern;
    this.severity = options?.severity;
    Object.setPrototypeOf(this, InjectionDetectedError.prototype);
  }
}

/**
 * Raised when tool access is denied (403)
 */
export class ToolAccessDeniedError extends GatewayOpsError {
  readonly mcpServer?: string;
  readonly toolName?: string;
  readonly requiresApproval: boolean;

  constructor(
    message = 'Tool access denied',
    options?: {
      mcpServer?: string;
      toolName?: string;
      requiresApproval?: boolean;
      details?: Record<string, unknown>;
    }
  ) {
    super(message, { code: 'tool_access_denied', statusCode: 403, details: options?.details });
    this.name = 'ToolAccessDeniedError';
    this.mcpServer = options?.mcpServer;
    this.toolName = options?.toolName;
    this.requiresApproval = options?.requiresApproval ?? false;
    Object.setPrototypeOf(this, ToolAccessDeniedError.prototype);
  }
}

/**
 * Raised when the server returns an error (5xx)
 */
export class ServerError extends GatewayOpsError {
  constructor(
    message = 'Server error',
    options?: { code?: string; details?: Record<string, unknown> }
  ) {
    super(message, { code: options?.code ?? 'server_error', statusCode: 500, details: options?.details });
    this.name = 'ServerError';
    Object.setPrototypeOf(this, ServerError.prototype);
  }
}

/**
 * Raised when a request times out
 */
export class TimeoutError extends GatewayOpsError {
  readonly timeoutMs?: number;

  constructor(
    message = 'Request timed out',
    options?: { timeoutMs?: number; details?: Record<string, unknown> }
  ) {
    super(message, { code: 'timeout', statusCode: 408, details: options?.details });
    this.name = 'TimeoutError';
    this.timeoutMs = options?.timeoutMs;
    Object.setPrototypeOf(this, TimeoutError.prototype);
  }
}

/**
 * Raised when a network error occurs
 */
export class NetworkError extends GatewayOpsError {
  constructor(message = 'Network error', details?: Record<string, unknown>) {
    super(message, { code: 'network_error', details });
    this.name = 'NetworkError';
    Object.setPrototypeOf(this, NetworkError.prototype);
  }
}
