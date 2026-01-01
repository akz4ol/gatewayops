# GatewayOps Python SDK

Official Python SDK for the GatewayOps MCP Gateway.

## Installation

```bash
pip install gatewayops
```

## Quick Start

```python
from gatewayops import GatewayOps

# Initialize the client
gw = GatewayOps(api_key="gwo_prd_...")

# Call an MCP tool
result = gw.mcp("filesystem").tools.call("read_file", path="/data.csv")
print(result.content)

# List available tools
tools = gw.mcp("filesystem").tools.list()
for tool in tools:
    print(f"{tool.name}: {tool.description}")
```

## Features

- **MCP Operations**: Call tools, read resources, get prompts
- **Tracing**: Distributed tracing with trace context
- **Cost Tracking**: Monitor usage and costs
- **Type Safety**: Full type hints with Pydantic models
- **Error Handling**: Detailed exception hierarchy
- **Retry Logic**: Built-in retries with exponential backoff

## MCP Operations

### Tools

```python
# List tools
tools = gw.mcp("filesystem").tools.list()

# Call a tool
result = gw.mcp("filesystem").tools.call(
    "read_file",
    path="/data/input.csv"
)

# Check for errors
if result.is_error:
    print(f"Error: {result.content}")
else:
    print(result.content)
```

### Resources

```python
# List resources
resources = gw.mcp("database").resources.list()

# Read a resource
content = gw.mcp("database").resources.read("db://users/schema")
print(content.text)
```

### Prompts

```python
# List prompts
prompts = gw.mcp("assistant").prompts.list()

# Get a prompt with arguments
messages = gw.mcp("assistant").prompts.get(
    "summarize",
    arguments={"length": "short"}
)
```

## Tracing

Use trace contexts to correlate multiple operations:

```python
with gw.trace("data-pipeline") as trace:
    # All operations in this block share the trace ID
    data = gw.mcp("filesystem").tools.call("read_file", path="/input.csv")
    result = gw.mcp("processor").tools.call("transform", data=data.content)
    gw.mcp("filesystem").tools.call("write_file", path="/output.csv", content=result.content)

    print(f"Trace ID: {trace.trace_id}")
```

### Viewing Traces

```python
# List recent traces
page = gw.traces.list(limit=10)
for trace in page.traces:
    print(f"{trace.id}: {trace.operation} - {trace.status}")

# Get trace details
trace = gw.traces.get("trace-id-here")
for span in trace.spans:
    print(f"  {span.name}: {span.duration_ms}ms")
```

## Cost Tracking

```python
# Get monthly cost summary
summary = gw.costs.summary(period="month")
print(f"Total cost: ${summary.total_cost:.2f}")
print(f"Request count: {summary.request_count}")

# Costs by MCP server
by_server = gw.costs.by_server()
for breakdown in by_server.by_server:
    print(f"{breakdown.value}: ${breakdown.cost:.2f}")

# Costs by team
by_team = gw.costs.by_team()
for breakdown in by_team.by_team:
    print(f"{breakdown.value}: ${breakdown.cost:.2f}")
```

## Error Handling

The SDK provides specific exceptions for different error types:

```python
from gatewayops import (
    GatewayOpsError,
    AuthenticationError,
    RateLimitError,
    NotFoundError,
    ValidationError,
    InjectionDetectedError,
    ToolAccessDeniedError,
)

try:
    result = gw.mcp("filesystem").tools.call("read_file", path="/secret.txt")
except AuthenticationError:
    print("Invalid API key")
except RateLimitError as e:
    print(f"Rate limited. Retry after {e.retry_after} seconds")
except ToolAccessDeniedError as e:
    if e.requires_approval:
        print(f"Tool {e.tool_name} requires approval")
    else:
        print(f"Access denied to {e.tool_name}")
except InjectionDetectedError as e:
    print(f"Prompt injection detected: {e.pattern}")
except NotFoundError:
    print("MCP server or tool not found")
except ValidationError as e:
    print(f"Validation error: {e.message}")
except GatewayOpsError as e:
    print(f"Error [{e.code}]: {e.message}")
```

## Configuration

### Custom Base URL

```python
gw = GatewayOps(
    api_key="gwo_prd_...",
    base_url="https://gateway.internal.company.com"
)
```

### Timeout

```python
gw = GatewayOps(
    api_key="gwo_prd_...",
    timeout=60.0  # 60 seconds
)
```

### Retries

```python
gw = GatewayOps(
    api_key="gwo_prd_...",
    max_retries=5  # Default is 3
)
```

## Context Manager

Use the client as a context manager for proper cleanup:

```python
with GatewayOps(api_key="gwo_prd_...") as gw:
    result = gw.mcp("filesystem").tools.call("read_file", path="/data.csv")
# HTTP client is automatically closed
```

## Async Support

For async applications, use the async client:

```python
from gatewayops import AsyncGatewayOps

async def main():
    async with AsyncGatewayOps(api_key="gwo_prd_...") as gw:
        result = await gw.mcp("filesystem").tools.call("read_file", path="/data.csv")
        print(result.content)
```

## Type Reference

### ToolCallResult

```python
@dataclass
class ToolCallResult:
    content: Any           # Tool output
    is_error: bool         # Whether the call failed
    metadata: dict | None  # Additional metadata
    trace_id: str | None   # Trace ID for this call
    span_id: str | None    # Span ID within the trace
    duration_ms: int | None  # Execution time
    cost: float | None     # Cost of this call
```

### Trace

```python
@dataclass
class Trace:
    id: str
    org_id: str
    mcp_server: str
    operation: str
    status: str  # "success", "error"
    start_time: datetime
    end_time: datetime | None
    duration_ms: int | None
    spans: list[Span] | None
    error_message: str | None
    cost: float | None
```

### CostSummary

```python
@dataclass
class CostSummary:
    total_cost: float
    period_start: datetime
    period_end: datetime
    request_count: int
    by_server: list[CostBreakdown] | None
    by_team: list[CostBreakdown] | None
    by_tool: list[CostBreakdown] | None
```

## Environment Variables

The SDK supports configuration via environment variables:

```bash
export GATEWAYOPS_API_KEY="gwo_prd_..."
export GATEWAYOPS_BASE_URL="https://api.gatewayops.com"
```

```python
import os
from gatewayops import GatewayOps

gw = GatewayOps(api_key=os.environ["GATEWAYOPS_API_KEY"])
```

## Requirements

- Python 3.8+
- httpx >= 0.25.0
- pydantic >= 2.0.0
- tenacity >= 8.0.0

## License

MIT License - see LICENSE file for details.
