/**
 * Type exports for GatewayOps SDK
 */

export * from './mcp';
export * from './traces';
export * from './costs';

/**
 * API Key information
 */
export interface APIKey {
  id: string;
  name: string;
  keyPrefix: string;
  environment: 'production' | 'sandbox';
  permissions: string;
  rateLimitRpm: number;
  createdAt: Date;
  lastUsedAt?: Date;
  expiresAt?: Date;
}

/**
 * Options for initializing the GatewayOps client
 */
export interface GatewayOpsOptions {
  /**
   * GatewayOps API key (e.g., "gwo_prd_...")
   */
  apiKey: string;

  /**
   * Base URL for the API
   * @default "https://api.gatewayops.com"
   */
  baseUrl?: string;

  /**
   * Request timeout in milliseconds
   * @default 30000
   */
  timeout?: number;

  /**
   * Maximum number of retries for failed requests
   * @default 3
   */
  maxRetries?: number;
}

/**
 * Error response from the API
 */
export interface APIError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

/**
 * API response wrapper
 */
export interface APIResponse<T> {
  data?: T;
  error?: APIError;
}
