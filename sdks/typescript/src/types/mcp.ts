/**
 * MCP (Model Context Protocol) types
 */

/**
 * Tool definition from an MCP server
 */
export interface ToolDefinition {
  name: string;
  description?: string;
  inputSchema?: Record<string, unknown>;
}

/**
 * Result of calling a tool
 */
export interface ToolCallResult {
  content: unknown;
  isError: boolean;
  metadata?: Record<string, unknown>;
  traceId?: string;
  spanId?: string;
  durationMs?: number;
  cost?: number;
}

/**
 * MCP resource definition
 */
export interface Resource {
  uri: string;
  name: string;
  description?: string;
  mimeType?: string;
}

/**
 * Content of an MCP resource
 */
export interface ResourceContent {
  uri: string;
  mimeType?: string;
  text?: string;
  blob?: string; // Base64 encoded
}

/**
 * MCP prompt definition
 */
export interface Prompt {
  name: string;
  description?: string;
  arguments?: PromptArgument[];
}

/**
 * Prompt argument definition
 */
export interface PromptArgument {
  name: string;
  description?: string;
  required?: boolean;
}

/**
 * Message in a prompt response
 */
export interface PromptMessage {
  role: 'user' | 'assistant' | 'system';
  content: unknown;
}
