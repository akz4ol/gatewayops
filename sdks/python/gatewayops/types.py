"""GatewayOps SDK types and models."""

from datetime import datetime
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field


class ToolDefinition(BaseModel):
    """Represents an MCP tool definition."""

    name: str
    description: Optional[str] = None
    input_schema: Optional[Dict[str, Any]] = Field(default=None, alias="inputSchema")

    class Config:
        populate_by_name = True


class ToolCallResult(BaseModel):
    """Represents the result of a tool call."""

    content: Any
    is_error: bool = Field(default=False, alias="isError")
    metadata: Optional[Dict[str, Any]] = None
    trace_id: Optional[str] = Field(default=None, alias="traceId")
    span_id: Optional[str] = Field(default=None, alias="spanId")
    duration_ms: Optional[int] = Field(default=None, alias="durationMs")
    cost: Optional[float] = None

    class Config:
        populate_by_name = True


class Resource(BaseModel):
    """Represents an MCP resource."""

    uri: str
    name: str
    description: Optional[str] = None
    mime_type: Optional[str] = Field(default=None, alias="mimeType")

    class Config:
        populate_by_name = True


class ResourceContent(BaseModel):
    """Represents the content of an MCP resource."""

    uri: str
    mime_type: Optional[str] = Field(default=None, alias="mimeType")
    text: Optional[str] = None
    blob: Optional[str] = None  # Base64 encoded

    class Config:
        populate_by_name = True


class Prompt(BaseModel):
    """Represents an MCP prompt."""

    name: str
    description: Optional[str] = None
    arguments: Optional[List[Dict[str, Any]]] = None

    class Config:
        populate_by_name = True


class PromptMessage(BaseModel):
    """Represents a message in an MCP prompt."""

    role: str
    content: Any


class Span(BaseModel):
    """Represents a trace span."""

    id: str
    trace_id: str = Field(alias="traceId")
    parent_span_id: Optional[str] = Field(default=None, alias="parentSpanId")
    name: str
    kind: str
    status: str
    start_time: datetime = Field(alias="startTime")
    end_time: Optional[datetime] = Field(default=None, alias="endTime")
    duration_ms: Optional[int] = Field(default=None, alias="durationMs")
    attributes: Optional[Dict[str, Any]] = None
    events: Optional[List[Dict[str, Any]]] = None

    class Config:
        populate_by_name = True


class Trace(BaseModel):
    """Represents a distributed trace."""

    id: str
    org_id: str = Field(alias="orgId")
    api_key_id: Optional[str] = Field(default=None, alias="apiKeyId")
    mcp_server: str = Field(alias="mcpServer")
    operation: str
    status: str
    start_time: datetime = Field(alias="startTime")
    end_time: Optional[datetime] = Field(default=None, alias="endTime")
    duration_ms: Optional[int] = Field(default=None, alias="durationMs")
    spans: Optional[List[Span]] = None
    error_message: Optional[str] = Field(default=None, alias="errorMessage")
    cost: Optional[float] = None

    class Config:
        populate_by_name = True


class TracePage(BaseModel):
    """Represents a paginated list of traces."""

    traces: List[Trace]
    total: int
    limit: int
    offset: int
    has_more: bool = Field(alias="hasMore")

    class Config:
        populate_by_name = True


class CostBreakdown(BaseModel):
    """Represents cost breakdown by dimension."""

    dimension: str
    value: str
    cost: float
    request_count: int = Field(alias="requestCount")

    class Config:
        populate_by_name = True


class CostSummary(BaseModel):
    """Represents a cost summary."""

    total_cost: float = Field(alias="totalCost")
    period_start: datetime = Field(alias="periodStart")
    period_end: datetime = Field(alias="periodEnd")
    request_count: int = Field(alias="requestCount")
    by_server: Optional[List[CostBreakdown]] = Field(default=None, alias="byServer")
    by_team: Optional[List[CostBreakdown]] = Field(default=None, alias="byTeam")
    by_tool: Optional[List[CostBreakdown]] = Field(default=None, alias="byTool")

    class Config:
        populate_by_name = True


class APIKey(BaseModel):
    """Represents an API key."""

    id: str
    name: str
    key_prefix: str = Field(alias="keyPrefix")
    environment: str
    permissions: str
    rate_limit_rpm: int = Field(alias="rateLimitRpm")
    created_at: datetime = Field(alias="createdAt")
    last_used_at: Optional[datetime] = Field(default=None, alias="lastUsedAt")
    expires_at: Optional[datetime] = Field(default=None, alias="expiresAt")

    class Config:
        populate_by_name = True


class TraceFilter(BaseModel):
    """Filter for listing traces."""

    mcp_server: Optional[str] = Field(default=None, alias="mcpServer")
    operation: Optional[str] = None
    status: Optional[str] = None
    start_time: Optional[datetime] = Field(default=None, alias="startTime")
    end_time: Optional[datetime] = Field(default=None, alias="endTime")
    limit: int = 50
    offset: int = 0

    class Config:
        populate_by_name = True
