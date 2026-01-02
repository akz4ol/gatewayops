"""
GatewayOps Python SDK

Official Python SDK for the GatewayOps MCP Gateway.

Example:
    >>> from gatewayops import GatewayOps
    >>> gw = GatewayOps(api_key="gwo_prd_...")
    >>> result = gw.mcp("filesystem").tools.call("read_file", path="/data.csv")
"""

__version__ = "0.1.2"

from gatewayops.client import GatewayOps
from gatewayops.exceptions import (
    GatewayOpsError,
    AuthenticationError,
    RateLimitError,
    NotFoundError,
    ValidationError,
    InjectionDetectedError,
    ToolAccessDeniedError,
)
from gatewayops.types import (
    ToolCallResult,
    ToolDefinition,
    Resource,
    Prompt,
    Trace,
    Span,
    CostSummary,
)
__all__ = [
    "GatewayOps",
    "GatewayOpsError",
    "AuthenticationError",
    "RateLimitError",
    "NotFoundError",
    "ValidationError",
    "InjectionDetectedError",
    "ToolAccessDeniedError",
    "ToolCallResult",
    "ToolDefinition",
    "Resource",
    "Prompt",
    "Trace",
    "Span",
    "CostSummary",
]
