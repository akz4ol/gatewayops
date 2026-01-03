# GatewayOps Agent Platform Integration Specification

## Executive Summary

This specification defines how GatewayOps MCP Gateway integrates with AI agent platforms (like https://frontend-akz4ol.vercel.app/) and supports industry-standard protocols for agentic AI systems.

**Target:** Enterprise-grade MCP gateway supporting multi-provider, multi-protocol AI agent deployments.

---

## 1. Protocol Support Matrix

### 1.1 Core Protocols

| Protocol | Version | Status | Transport |
|----------|---------|--------|-----------|
| MCP (Model Context Protocol) | 1.0 | ✅ Supported | HTTP, WebSocket, stdio |
| OpenAI Agents SDK | 2025.x | 🔄 Planned | HTTP |
| LangChain MCP Adapters | 0.3+ | 🔄 Planned | HTTP |
| JSON-RPC 2.0 | 2.0 | ✅ Supported | HTTP, WebSocket |
| OpenTelemetry (OTLP) | 1.x | ✅ Supported | gRPC, HTTP |

### 1.2 Transport Protocols

| Transport | Use Case | Status |
|-----------|----------|--------|
| HTTP/REST | Stateless tool calls | ✅ Supported |
| WebSocket | Real-time bidirectional | 🔄 Planned |
| Server-Sent Events (SSE) | Streaming responses | 🔄 Planned |
| stdio | Local MCP servers | ✅ Supported (proxy) |
| gRPC | High-performance | 🔜 Future |

---

## 2. Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Agent Platform Frontend                          │
│                  (frontend-akz4ol.vercel.app)                          │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │ HTTPS/WebSocket
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         GatewayOps MCP Gateway                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Auth &    │  │   Rate      │  │   Safety    │  │  Routing &  │   │
│  │    RBAC     │  │   Limiting  │  │   Layer     │  │  Load Bal   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Protocol   │  │   Tracing   │  │   Cost      │  │   Caching   │   │
│  │  Adapters   │  │   & OTEL    │  │   Tracking  │  │   Layer     │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        ▼                         ▼                         ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│  MCP Servers  │       │  LLM Providers│       │  External     │
│  (filesystem, │       │  (OpenAI,     │       │  Services     │
│   github,     │       │   Anthropic,  │       │  (Slack,      │
│   database)   │       │   local)      │       │   GitHub)     │
└───────────────┘       └───────────────┘       └───────────────┘
```

---

## 3. API Specification

### 3.1 Agent Platform Connection Endpoint

**Purpose:** Allow agent platforms to register and establish connections.

```
POST /v1/agents/connect
```

**Request:**
```json
{
  "agent_id": "agent_abc123",
  "platform": "frontend-akz4ol",
  "capabilities": ["tool_calling", "streaming", "multi_turn"],
  "transport": "websocket",
  "callback_url": "https://frontend-akz4ol.vercel.app/api/webhook",
  "metadata": {
    "user_id": "user_123",
    "session_id": "sess_456"
  }
}
```

**Response:**
```json
{
  "connection_id": "conn_xyz789",
  "gateway_url": "wss://api.gatewayops.com/v1/agents/conn_xyz789",
  "available_servers": [
    {
      "name": "filesystem",
      "tools": 12,
      "resources": 5
    },
    {
      "name": "github",
      "tools": 8,
      "resources": 3
    }
  ],
  "rate_limits": {
    "requests_per_minute": 1000,
    "tokens_per_minute": 100000
  }
}
```

### 3.2 Universal Tool Execution

**Purpose:** Execute tools across any registered MCP server with protocol translation.

```
POST /v1/execute
```

**Request:**
```json
{
  "connection_id": "conn_xyz789",
  "calls": [
    {
      "id": "call_001",
      "server": "filesystem",
      "tool": "read_file",
      "arguments": {
        "path": "/src/main.ts"
      }
    },
    {
      "id": "call_002", 
      "server": "github",
      "tool": "create_issue",
      "arguments": {
        "repo": "akz4ol/gatewayops",
        "title": "Bug fix",
        "body": "Description here"
      }
    }
  ],
  "execution_mode": "parallel",
  "timeout_ms": 30000
}
```

**Response:**
```json
{
  "results": [
    {
      "id": "call_001",
      "status": "success",
      "content": [
        {
          "type": "text",
          "text": "// File contents..."
        }
      ],
      "duration_ms": 45,
      "cost": 0.0001
    },
    {
      "id": "call_002",
      "status": "success",
      "content": [
        {
          "type": "text",
          "text": "Issue #123 created"
        }
      ],
      "duration_ms": 230,
      "cost": 0.0003
    }
  ],
  "trace_id": "tr_abc123",
  "total_cost": 0.0004
}
```

### 3.3 Streaming Tool Execution (SSE)

**Purpose:** Stream tool results for long-running operations.

```
POST /v1/execute/stream
Accept: text/event-stream
```

**Response Stream:**
```
event: start
data: {"call_id": "call_001", "server": "filesystem", "tool": "read_file"}

event: progress
data: {"call_id": "call_001", "progress": 0.5, "message": "Reading file..."}

event: chunk
data: {"call_id": "call_001", "content": "partial content..."}

event: complete
data: {"call_id": "call_001", "status": "success", "duration_ms": 145}

event: done
data: {"trace_id": "tr_abc123", "total_cost": 0.0002}
```

### 3.4 WebSocket Protocol

**Purpose:** Real-time bidirectional communication for interactive agents.

```
wss://api.gatewayops.com/v1/agents/{connection_id}
```

**Client → Server Messages:**

```json
// Tool call request
{
  "type": "tool_call",
  "id": "msg_001",
  "payload": {
    "server": "filesystem",
    "tool": "write_file",
    "arguments": {"path": "/tmp/test.txt", "content": "Hello"}
  }
}

// Cancel request
{
  "type": "cancel",
  "id": "msg_001"
}

// Ping
{
  "type": "ping"
}
```

**Server → Client Messages:**

```json
// Tool result
{
  "type": "tool_result",
  "id": "msg_001",
  "payload": {
    "status": "success",
    "content": [{"type": "text", "text": "File written"}]
  }
}

// Progress update
{
  "type": "progress",
  "id": "msg_001",
  "payload": {
    "progress": 0.75,
    "message": "Writing to disk..."
  }
}

// Error
{
  "type": "error",
  "id": "msg_001",
  "payload": {
    "code": "permission_denied",
    "message": "Tool requires approval"
  }
}

// Pong
{
  "type": "pong"
}
```

---

## 4. OpenAI Agents SDK Compatibility

### 4.1 MCP Server Registration for OpenAI

GatewayOps acts as an MCP server that OpenAI Agents SDK can connect to.

**Configuration for OpenAI Agents:**

```python
from openai import OpenAI
from openai.agents import Agent

client = OpenAI()

# Register GatewayOps as MCP server
agent = Agent(
    name="coding-assistant",
    model="gpt-4o",
    mcp_servers=[
        {
            "type": "url",
            "url": "https://api.gatewayops.com/v1/mcp",
            "headers": {
                "Authorization": "Bearer gwo_prd_xxx"
            }
        }
    ]
)

# Agent can now use all tools from all MCP servers behind GatewayOps
response = agent.run("Read the file /src/main.ts and create a GitHub issue")
```

### 4.2 Tool Discovery Endpoint

**OpenAI-compatible tool listing:**

```
GET /v1/mcp/tools
```

**Response (OpenAI format):**
```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "filesystem__read_file",
        "description": "Read contents of a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file"
            }
          },
          "required": ["path"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "github__create_issue",
        "description": "Create a GitHub issue",
        "parameters": {
          "type": "object",
          "properties": {
            "repo": {"type": "string"},
            "title": {"type": "string"},
            "body": {"type": "string"}
          },
          "required": ["repo", "title"]
        }
      }
    }
  ]
}
```

---

## 5. LangChain Integration

### 5.1 LangChain MCP Adapter Support

```python
from langchain_mcp_adapters import MCPToolkit
from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_openai import ChatOpenAI

# Connect LangChain to GatewayOps
toolkit = MCPToolkit(
    server_url="https://api.gatewayops.com/v1/mcp",
    headers={"Authorization": "Bearer gwo_prd_xxx"}
)

tools = toolkit.get_tools()

# Create agent with GatewayOps tools
llm = ChatOpenAI(model="gpt-4o")
agent = create_tool_calling_agent(llm, tools, prompt)
agent_executor = AgentExecutor(agent=agent, tools=tools)

result = agent_executor.invoke({
    "input": "Read /src/main.ts and summarize it"
})
```

### 5.2 LangGraph Support

```python
from langgraph.prebuilt import create_react_agent
from langchain_mcp_adapters import MCPToolkit

toolkit = MCPToolkit(
    server_url="https://api.gatewayops.com/v1/mcp",
    headers={"Authorization": "Bearer gwo_prd_xxx"}
)

graph = create_react_agent(
    model=ChatOpenAI(model="gpt-4o"),
    tools=toolkit.get_tools()
)
```

---

## 6. Security Layer

### 6.1 Prompt Injection Detection

**Enhanced detection for agent workloads:**

```
POST /v1/safety/analyze
```

**Request:**
```json
{
  "content": "Ignore previous instructions and...",
  "context": {
    "source": "user_input",
    "tool_chain": ["filesystem__read_file", "github__create_issue"]
  }
}
```

**Response:**
```json
{
  "safe": false,
  "detections": [
    {
      "type": "prompt_injection",
      "severity": "high",
      "pattern": "instruction_override",
      "confidence": 0.95,
      "recommendation": "block"
    }
  ],
  "sanitized_content": null
}
```

### 6.2 Tool Permission Scopes

```json
{
  "scopes": {
    "filesystem": {
      "read": ["/**/*.ts", "/**/*.md"],
      "write": ["/tmp/**"],
      "deny": ["/etc/**", "/**/.env"]
    },
    "github": {
      "repos": ["akz4ol/*"],
      "actions": ["read", "create_issue", "create_pr"],
      "deny_actions": ["delete_repo", "admin"]
    },
    "database": {
      "operations": ["select"],
      "deny_operations": ["drop", "truncate", "delete"]
    }
  }
}
```

### 6.3 Tool Chaining Analysis

Detect dangerous tool combinations:

```json
{
  "chain_rules": [
    {
      "name": "data_exfiltration",
      "pattern": ["*__read_*", "slack__send_message"],
      "action": "require_approval"
    },
    {
      "name": "destructive_sequence",
      "pattern": ["filesystem__read_*", "filesystem__delete_*"],
      "action": "block"
    }
  ]
}
```

---

## 7. Observability

### 7.1 OpenTelemetry Integration

**Trace Context Propagation:**

```
Headers:
  traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
  tracestate: gatewayops=conn_xyz789
```

**Span Attributes:**

```json
{
  "gatewayops.connection_id": "conn_xyz789",
  "gatewayops.agent_platform": "frontend-akz4ol",
  "gatewayops.mcp_server": "filesystem",
  "gatewayops.tool_name": "read_file",
  "gatewayops.user_id": "user_123",
  "gatewayops.cost_usd": 0.0001
}
```

### 7.2 Real-time Metrics

```
GET /v1/metrics/realtime
```

**Response:**
```json
{
  "connections": {
    "active": 145,
    "by_platform": {
      "frontend-akz4ol": 89,
      "langchain": 34,
      "openai-agents": 22
    }
  },
  "requests": {
    "per_second": 234.5,
    "latency_p50_ms": 45,
    "latency_p99_ms": 230
  },
  "tools": {
    "calls_per_minute": 5420,
    "error_rate": 0.02,
    "top_tools": [
      {"name": "filesystem__read_file", "calls": 2340},
      {"name": "github__search_code", "calls": 890}
    ]
  }
}
```

---

## 8. Agent Platform SDK

### 8.1 TypeScript SDK for Agent Platforms

```typescript
import { GatewayOps, AgentConnection } from '@gatewayops/sdk';

// Initialize with agent platform config
const gw = new GatewayOps({
  apiKey: 'gwo_prd_xxx',
  platform: 'frontend-akz4ol'
});

// Establish WebSocket connection
const connection = await gw.agents.connect({
  capabilities: ['tool_calling', 'streaming'],
  transport: 'websocket'
});

// Execute tools with streaming
const stream = connection.execute({
  server: 'filesystem',
  tool: 'read_file',
  arguments: { path: '/src/main.ts' },
  stream: true
});

for await (const event of stream) {
  if (event.type === 'chunk') {
    console.log(event.content);
  }
}

// Multi-tool execution
const results = await connection.executeBatch([
  { server: 'filesystem', tool: 'read_file', arguments: { path: '/src/a.ts' } },
  { server: 'filesystem', tool: 'read_file', arguments: { path: '/src/b.ts' } },
  { server: 'github', tool: 'get_repo', arguments: { repo: 'akz4ol/gatewayops' } }
], { parallel: true });
```

### 8.2 React Hooks

```typescript
import { useGatewayOps, useToolCall } from '@gatewayops/react';

function CodeAssistant() {
  const { connection, status } = useGatewayOps({
    apiKey: process.env.GATEWAYOPS_API_KEY
  });

  const { execute, result, loading, error } = useToolCall(connection);

  const handleReadFile = async (path: string) => {
    await execute({
      server: 'filesystem',
      tool: 'read_file',
      arguments: { path }
    });
  };

  return (
    <div>
      <button onClick={() => handleReadFile('/src/main.ts')}>
        Read File
      </button>
      {loading && <Spinner />}
      {result && <CodeBlock content={result.content} />}
      {error && <ErrorMessage error={error} />}
    </div>
  );
}
```

---

## 9. Configuration

### 9.1 Agent Platform Registration

```yaml
# gatewayops.yaml
platforms:
  frontend-akz4ol:
    display_name: "Agent Platform"
    callback_url: "https://frontend-akz4ol.vercel.app/api/webhook"
    allowed_origins:
      - "https://frontend-akz4ol.vercel.app"
    rate_limits:
      requests_per_minute: 1000
      concurrent_connections: 100
    default_servers:
      - filesystem
      - github
      - database
    security:
      require_approval_for:
        - "*.delete_*"
        - "*.write_*"
      block_patterns:
        - "database.drop_*"

servers:
  filesystem:
    type: stdio
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"]
    
  github:
    type: http
    url: "https://mcp-github.example.com"
    auth:
      type: bearer
      token_env: GITHUB_MCP_TOKEN

  database:
    type: http  
    url: "https://mcp-postgres.example.com"
    auth:
      type: bearer
      token_env: DATABASE_MCP_TOKEN
```

---

## 10. Implementation Roadmap

### Phase 1: Core Agent Support (2 weeks)
- [ ] WebSocket transport implementation
- [ ] Agent connection management
- [ ] Batch tool execution
- [ ] Real-time metrics endpoint

### Phase 2: Protocol Adapters (2 weeks)
- [ ] OpenAI Agents SDK compatibility layer
- [ ] LangChain MCP adapter support
- [ ] Tool name translation (server__tool format)
- [ ] SSE streaming endpoint

### Phase 3: Security Enhancements (1 week)
- [ ] Tool chain analysis
- [ ] Enhanced prompt injection detection
- [ ] Scope-based permissions
- [ ] Audit logging for agent sessions

### Phase 4: SDKs & React Integration (1 week)
- [ ] TypeScript SDK agent extensions
- [ ] React hooks package (@gatewayops/react)
- [ ] Python SDK agent extensions
- [ ] Example agent platform integration

---

## 11. References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [OpenAI Agents SDK MCP Support](https://openai.github.io/openai-agents-python/mcp/)
- [LangChain MCP Adapters](https://docs.langchain.com/oss/python/langchain/mcp)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/)

---

*Version: 1.0.0*
*Last Updated: 2026-01-03*
*Author: GatewayOps Engineering*
