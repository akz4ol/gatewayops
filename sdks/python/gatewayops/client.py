"""GatewayOps SDK client."""

from typing import Any, Dict, List, Optional, Union
from contextlib import contextmanager
import httpx
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

from gatewayops.exceptions import (
    GatewayOpsError,
    AuthenticationError,
    RateLimitError,
    NotFoundError,
    ValidationError,
    InjectionDetectedError,
    ToolAccessDeniedError,
    ServerError,
    NetworkError,
)
from gatewayops.types import (
    ToolCallResult,
    ToolDefinition,
    Resource,
    ResourceContent,
    Prompt,
    PromptMessage,
    Trace,
    TracePage,
    CostSummary,
    TraceFilter,
)


class GatewayOps:
    """
    GatewayOps SDK client.

    Example:
        >>> gw = GatewayOps(api_key="gwo_prd_...")
        >>> result = gw.mcp("filesystem").tools.call("read_file", path="/data.csv")
    """

    DEFAULT_BASE_URL = "https://api.gatewayops.com"
    DEFAULT_TIMEOUT = 30.0

    def __init__(
        self,
        api_key: str,
        base_url: Optional[str] = None,
        timeout: Optional[float] = None,
        max_retries: int = 3,
    ):
        """
        Initialize the GatewayOps client.

        Args:
            api_key: GatewayOps API key (e.g., "gwo_prd_...")
            base_url: Base URL for the API (default: https://api.gatewayops.com)
            timeout: Request timeout in seconds (default: 30)
            max_retries: Maximum number of retries for failed requests (default: 3)
        """
        self.api_key = api_key
        self.base_url = (base_url or self.DEFAULT_BASE_URL).rstrip("/")
        self.timeout = timeout or self.DEFAULT_TIMEOUT
        self.max_retries = max_retries
        self._trace_context: Optional[str] = None

        self._client = httpx.Client(
            base_url=self.base_url,
            headers={
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/json",
                "User-Agent": "gatewayops-python/0.1.0",
            },
            timeout=self.timeout,
        )

    def __enter__(self) -> "GatewayOps":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def close(self) -> None:
        """Close the HTTP client."""
        self._client.close()

    def mcp(self, server: str) -> "MCPClient":
        """
        Get an MCP client for a specific server.

        Args:
            server: Name of the MCP server

        Returns:
            MCPClient for the specified server
        """
        return MCPClient(self, server)

    @contextmanager
    def trace(self, name: str):
        """
        Create a tracing context.

        Args:
            name: Name for the trace

        Yields:
            Trace context
        """
        # Generate a trace ID or use existing
        import uuid
        trace_id = str(uuid.uuid4())
        old_context = self._trace_context
        self._trace_context = trace_id
        try:
            yield TraceContext(trace_id, name)
        finally:
            self._trace_context = old_context

    @property
    def traces(self) -> "TracesClient":
        """Get the traces client."""
        return TracesClient(self)

    @property
    def costs(self) -> "CostsClient":
        """Get the costs client."""
        return CostsClient(self)

    def _request(
        self,
        method: str,
        path: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Make an HTTP request to the API."""
        headers = {}
        if self._trace_context:
            headers["X-Trace-ID"] = self._trace_context

        try:
            response = self._client.request(
                method=method,
                url=path,
                json=data,
                params=params,
                headers=headers,
            )
            return self._handle_response(response)
        except httpx.TimeoutException as e:
            raise NetworkError(f"Request timed out: {e}")
        except httpx.RequestError as e:
            raise NetworkError(f"Network error: {e}")

    def _handle_response(self, response: httpx.Response) -> Dict[str, Any]:
        """Handle the HTTP response."""
        try:
            data = response.json()
        except Exception:
            data = {}

        if response.status_code >= 400:
            self._raise_for_error(response.status_code, data)

        return data

    def _raise_for_error(self, status_code: int, data: Dict[str, Any]) -> None:
        """Raise an appropriate exception for an error response."""
        error_code = data.get("error", {}).get("code", "unknown")
        message = data.get("error", {}).get("message", "Unknown error")
        details = data.get("error", {}).get("details", {})

        if status_code == 401:
            raise AuthenticationError(message, code=error_code, details=details)
        elif status_code == 403:
            if error_code == "tool_access_denied":
                raise ToolAccessDeniedError(
                    message,
                    mcp_server=details.get("mcp_server"),
                    tool_name=details.get("tool_name"),
                    requires_approval=details.get("requires_approval", False),
                    details=details,
                )
            raise GatewayOpsError(message, code=error_code, status_code=403, details=details)
        elif status_code == 404:
            raise NotFoundError(message, details=details)
        elif status_code == 429:
            retry_after = None
            if "Retry-After" in details:
                retry_after = int(details["Retry-After"])
            raise RateLimitError(message, retry_after=retry_after, details=details)
        elif status_code == 400:
            if error_code == "injection_detected":
                raise InjectionDetectedError(
                    message,
                    pattern=details.get("pattern"),
                    severity=details.get("severity"),
                    details=details,
                )
            raise ValidationError(message, details=details)
        elif status_code >= 500:
            raise ServerError(message, code=error_code, details=details)
        else:
            raise GatewayOpsError(message, code=error_code, status_code=status_code, details=details)


class MCPClient:
    """Client for MCP operations on a specific server."""

    def __init__(self, client: GatewayOps, server: str):
        self._client = client
        self._server = server
        self._retries = client.max_retries

    @property
    def tools(self) -> "ToolsClient":
        """Get the tools client."""
        return ToolsClient(self._client, self._server)

    @property
    def resources(self) -> "ResourcesClient":
        """Get the resources client."""
        return ResourcesClient(self._client, self._server)

    @property
    def prompts(self) -> "PromptsClient":
        """Get the prompts client."""
        return PromptsClient(self._client, self._server)

    def with_retries(self, retries: int) -> "MCPClient":
        """Set the number of retries for this client."""
        new_client = MCPClient(self._client, self._server)
        new_client._retries = retries
        return new_client

    def with_trace(self, trace_name: str) -> "MCPClient":
        """Set a trace context for this client."""
        # This creates a new trace context on the parent client
        return self


class ToolsClient:
    """Client for MCP tool operations."""

    def __init__(self, client: GatewayOps, server: str):
        self._client = client
        self._server = server

    def list(self) -> List[ToolDefinition]:
        """List available tools."""
        response = self._client._request("POST", f"/v1/mcp/{self._server}/tools/list")
        tools_data = response.get("tools", [])
        return [ToolDefinition(**t) for t in tools_data]

    def call(self, tool: str, **arguments: Any) -> ToolCallResult:
        """
        Call a tool.

        Args:
            tool: Name of the tool to call
            **arguments: Tool arguments

        Returns:
            ToolCallResult with the result
        """
        response = self._client._request(
            "POST",
            f"/v1/mcp/{self._server}/tools/call",
            data={"tool": tool, "arguments": arguments},
        )
        return ToolCallResult(**response)


class ResourcesClient:
    """Client for MCP resource operations."""

    def __init__(self, client: GatewayOps, server: str):
        self._client = client
        self._server = server

    def list(self) -> List[Resource]:
        """List available resources."""
        response = self._client._request("POST", f"/v1/mcp/{self._server}/resources/list")
        resources_data = response.get("resources", [])
        return [Resource(**r) for r in resources_data]

    def read(self, uri: str) -> ResourceContent:
        """
        Read a resource.

        Args:
            uri: URI of the resource to read

        Returns:
            ResourceContent with the content
        """
        response = self._client._request(
            "POST",
            f"/v1/mcp/{self._server}/resources/read",
            data={"uri": uri},
        )
        return ResourceContent(**response)


class PromptsClient:
    """Client for MCP prompt operations."""

    def __init__(self, client: GatewayOps, server: str):
        self._client = client
        self._server = server

    def list(self) -> List[Prompt]:
        """List available prompts."""
        response = self._client._request("POST", f"/v1/mcp/{self._server}/prompts/list")
        prompts_data = response.get("prompts", [])
        return [Prompt(**p) for p in prompts_data]

    def get(self, name: str, arguments: Optional[Dict[str, Any]] = None) -> List[PromptMessage]:
        """
        Get a prompt.

        Args:
            name: Name of the prompt
            arguments: Prompt arguments

        Returns:
            List of prompt messages
        """
        response = self._client._request(
            "POST",
            f"/v1/mcp/{self._server}/prompts/get",
            data={"name": name, "arguments": arguments or {}},
        )
        messages_data = response.get("messages", [])
        return [PromptMessage(**m) for m in messages_data]


class TracesClient:
    """Client for trace operations."""

    def __init__(self, client: GatewayOps):
        self._client = client

    def list(
        self,
        mcp_server: Optional[str] = None,
        operation: Optional[str] = None,
        status: Optional[str] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> TracePage:
        """
        List traces.

        Args:
            mcp_server: Filter by MCP server
            operation: Filter by operation
            status: Filter by status
            limit: Maximum number of results
            offset: Offset for pagination

        Returns:
            TracePage with traces
        """
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        if mcp_server:
            params["mcp_server"] = mcp_server
        if operation:
            params["operation"] = operation
        if status:
            params["status"] = status

        response = self._client._request("GET", "/v1/traces", params=params)
        return TracePage(**response)

    def get(self, trace_id: str) -> Trace:
        """
        Get a specific trace.

        Args:
            trace_id: ID of the trace

        Returns:
            Trace details
        """
        response = self._client._request("GET", f"/v1/traces/{trace_id}")
        return Trace(**response)


class CostsClient:
    """Client for cost operations."""

    def __init__(self, client: GatewayOps):
        self._client = client

    def summary(
        self,
        period: str = "month",
        group_by: Optional[str] = None,
    ) -> CostSummary:
        """
        Get cost summary.

        Args:
            period: Time period (day, week, month)
            group_by: Group by dimension (server, team, tool)

        Returns:
            CostSummary with cost data
        """
        params: Dict[str, Any] = {"period": period}
        if group_by:
            params["group_by"] = group_by

        response = self._client._request("GET", "/v1/costs/summary", params=params)
        return CostSummary(**response)

    def by_server(self, period: str = "month") -> CostSummary:
        """Get costs grouped by MCP server."""
        return self.summary(period=period, group_by="server")

    def by_team(self, period: str = "month") -> CostSummary:
        """Get costs grouped by team."""
        return self.summary(period=period, group_by="team")


class TraceContext:
    """Context manager for tracing."""

    def __init__(self, trace_id: str, name: str):
        self.trace_id = trace_id
        self.name = name
