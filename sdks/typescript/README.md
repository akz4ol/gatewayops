# GatewayOps TypeScript SDK

Official TypeScript SDK for the GatewayOps MCP Gateway.

## Installation

```bash
npm install @gatewayops/sdk
# or
yarn add @gatewayops/sdk
# or
pnpm add @gatewayops/sdk
```

## Quick Start

```typescript
import { GatewayOps } from '@gatewayops/sdk';

// Initialize the client
const gw = new GatewayOps({ apiKey: 'gwo_prd_...' });

// Call an MCP tool
const result = await gw.mcp('filesystem').tools.call('read_file', { path: '/data.csv' });
console.log(result.content);

// List available tools
const tools = await gw.mcp('filesystem').tools.list();
for (const tool of tools) {
  console.log(`${tool.name}: ${tool.description}`);
}
```

## Features

- **MCP Operations**: Call tools, read resources, get prompts
- **Tracing**: Distributed tracing with trace context
- **Cost Tracking**: Monitor usage and costs
- **Full TypeScript**: Complete type definitions
- **Error Handling**: Detailed exception hierarchy
- **Retry Logic**: Built-in retries with exponential backoff

## MCP Operations

### Tools

```typescript
// List tools
const tools = await gw.mcp('filesystem').tools.list();

// Call a tool
const result = await gw.mcp('filesystem').tools.call('read_file', {
  path: '/data/input.csv',
});

// Check for errors
if (result.isError) {
  console.error(`Error: ${result.content}`);
} else {
  console.log(result.content);
}
```

### Resources

```typescript
// List resources
const resources = await gw.mcp('database').resources.list();

// Read a resource
const content = await gw.mcp('database').resources.read('db://users/schema');
console.log(content.text);
```

### Prompts

```typescript
// List prompts
const prompts = await gw.mcp('assistant').prompts.list();

// Get a prompt with arguments
const messages = await gw.mcp('assistant').prompts.get('summarize', {
  length: 'short',
});
```

## Tracing

Use trace contexts to correlate multiple operations:

```typescript
const trace = gw.trace('data-pipeline');

await trace.run(async () => {
  // All operations in this block share the trace ID
  const data = await gw.mcp('filesystem').tools.call('read_file', { path: '/input.csv' });
  const result = await gw.mcp('processor').tools.call('transform', { data: data.content });
  await gw.mcp('filesystem').tools.call('write_file', {
    path: '/output.csv',
    content: result.content,
  });

  console.log(`Trace ID: ${trace.traceId}`);
});
```

### Viewing Traces

```typescript
// List recent traces
const page = await gw.traces.list({ limit: 10 });
for (const trace of page.traces) {
  console.log(`${trace.id}: ${trace.operation} - ${trace.status}`);
}

// Get trace details
const trace = await gw.traces.get('trace-id-here');
for (const span of trace.spans ?? []) {
  console.log(`  ${span.name}: ${span.durationMs}ms`);
}
```

## Cost Tracking

```typescript
// Get monthly cost summary
const summary = await gw.costs.summary({ period: 'month' });
console.log(`Total cost: $${summary.totalCost.toFixed(2)}`);
console.log(`Request count: ${summary.requestCount}`);

// Costs by MCP server
const byServer = await gw.costs.byServer();
for (const breakdown of byServer.byServer ?? []) {
  console.log(`${breakdown.value}: $${breakdown.cost.toFixed(2)}`);
}

// Costs by team
const byTeam = await gw.costs.byTeam();
for (const breakdown of byTeam.byTeam ?? []) {
  console.log(`${breakdown.value}: $${breakdown.cost.toFixed(2)}`);
}
```

## Error Handling

The SDK provides specific error classes for different error types:

```typescript
import {
  GatewayOpsError,
  AuthenticationError,
  RateLimitError,
  NotFoundError,
  ValidationError,
  InjectionDetectedError,
  ToolAccessDeniedError,
} from '@gatewayops/sdk';

try {
  const result = await gw.mcp('filesystem').tools.call('read_file', { path: '/secret.txt' });
} catch (error) {
  if (error instanceof AuthenticationError) {
    console.error('Invalid API key');
  } else if (error instanceof RateLimitError) {
    console.error(`Rate limited. Retry after ${error.retryAfter} seconds`);
  } else if (error instanceof ToolAccessDeniedError) {
    if (error.requiresApproval) {
      console.error(`Tool ${error.toolName} requires approval`);
    } else {
      console.error(`Access denied to ${error.toolName}`);
    }
  } else if (error instanceof InjectionDetectedError) {
    console.error(`Prompt injection detected: ${error.pattern}`);
  } else if (error instanceof NotFoundError) {
    console.error('MCP server or tool not found');
  } else if (error instanceof ValidationError) {
    console.error(`Validation error: ${error.message}`);
  } else if (error instanceof GatewayOpsError) {
    console.error(`Error [${error.code}]: ${error.message}`);
  }
}
```

## Configuration

### Custom Base URL

```typescript
const gw = new GatewayOps({
  apiKey: 'gwo_prd_...',
  baseUrl: 'https://gateway.internal.company.com',
});
```

### Timeout

```typescript
const gw = new GatewayOps({
  apiKey: 'gwo_prd_...',
  timeout: 60000, // 60 seconds
});
```

### Retries

```typescript
const gw = new GatewayOps({
  apiKey: 'gwo_prd_...',
  maxRetries: 5, // Default is 3
});
```

## Type Reference

### ToolCallResult

```typescript
interface ToolCallResult {
  content: unknown; // Tool output
  isError: boolean; // Whether the call failed
  metadata?: Record<string, unknown>; // Additional metadata
  traceId?: string; // Trace ID for this call
  spanId?: string; // Span ID within the trace
  durationMs?: number; // Execution time
  cost?: number; // Cost of this call
}
```

### Trace

```typescript
interface Trace {
  id: string;
  orgId: string;
  mcpServer: string;
  operation: string;
  status: 'success' | 'error' | 'timeout';
  startTime: Date;
  endTime?: Date;
  durationMs?: number;
  spans?: Span[];
  errorMessage?: string;
  cost?: number;
}
```

### CostSummary

```typescript
interface CostSummary {
  totalCost: number;
  periodStart: Date;
  periodEnd: Date;
  requestCount: number;
  byServer?: CostBreakdown[];
  byTeam?: CostBreakdown[];
  byTool?: CostBreakdown[];
}
```

## Environment Variables

```typescript
const gw = new GatewayOps({
  apiKey: process.env.GATEWAYOPS_API_KEY!,
  baseUrl: process.env.GATEWAYOPS_BASE_URL,
});
```

## Requirements

- Node.js 18+
- TypeScript 5.0+ (for type definitions)

## License

MIT License - see LICENSE file for details.
