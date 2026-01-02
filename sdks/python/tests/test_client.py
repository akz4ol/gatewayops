"""Tests for GatewayOps SDK client."""

import pytest
from unittest.mock import Mock, patch, MagicMock
import httpx

from gatewayops import GatewayOps
from gatewayops.exceptions import (
    AuthenticationError,
    RateLimitError,
    NotFoundError,
    ValidationError,
    InjectionDetectedError,
    ToolAccessDeniedError,
    ServerError,
    NetworkError,
)


@pytest.fixture
def client():
    """Create a GatewayOps client for testing."""
    return GatewayOps(api_key="gwo_test_123", base_url="https://api.test.com")


@pytest.fixture
def mock_response():
    """Create a mock HTTP response."""
    def _mock_response(status_code: int, json_data: dict):
        response = Mock(spec=httpx.Response)
        response.status_code = status_code
        response.json.return_value = json_data
        return response
    return _mock_response


class TestGatewayOpsClient:
    """Tests for GatewayOps client initialization."""

    def test_default_base_url(self):
        """Client should use default base URL."""
        client = GatewayOps(api_key="test")
        assert client.base_url == "https://api.gatewayops.com"

    def test_custom_base_url(self):
        """Client should accept custom base URL."""
        client = GatewayOps(api_key="test", base_url="https://custom.api.com/")
        assert client.base_url == "https://custom.api.com"  # Trailing slash removed

    def test_default_timeout(self):
        """Client should use default timeout."""
        client = GatewayOps(api_key="test")
        assert client.timeout == 30.0

    def test_custom_timeout(self):
        """Client should accept custom timeout."""
        client = GatewayOps(api_key="test", timeout=60.0)
        assert client.timeout == 60.0

    def test_context_manager(self):
        """Client should work as context manager."""
        with GatewayOps(api_key="test") as client:
            assert client.api_key == "test"


class TestMCPClient:
    """Tests for MCP client."""

    def test_mcp_returns_client(self, client):
        """mcp() should return MCPClient instance."""
        mcp = client.mcp("filesystem")
        assert mcp._server == "filesystem"

    def test_mcp_tools_client(self, client):
        """MCPClient should provide tools client."""
        mcp = client.mcp("filesystem")
        tools = mcp.tools
        assert tools._server == "filesystem"

    def test_mcp_resources_client(self, client):
        """MCPClient should provide resources client."""
        mcp = client.mcp("filesystem")
        resources = mcp.resources
        assert resources._server == "filesystem"

    def test_mcp_prompts_client(self, client):
        """MCPClient should provide prompts client."""
        mcp = client.mcp("filesystem")
        prompts = mcp.prompts
        assert prompts._server == "filesystem"


class TestTracesClient:
    """Tests for traces client."""

    def test_traces_client_exists(self, client):
        """Client should provide traces client."""
        traces = client.traces
        assert traces is not None

    @patch.object(GatewayOps, "_request")
    def test_traces_list(self, mock_request, client):
        """traces.list() should call correct endpoint."""
        mock_request.return_value = {
            "traces": [],
            "total": 0,
            "limit": 50,
            "offset": 0,
        }

        result = client.traces.list(limit=10)

        mock_request.assert_called_once()
        call_args = mock_request.call_args
        assert call_args[0][0] == "GET"
        assert call_args[0][1] == "/v1/traces"
        assert result.traces == []

    @patch.object(GatewayOps, "_request")
    def test_traces_get(self, mock_request, client):
        """traces.get() should call correct endpoint."""
        mock_request.return_value = {
            "id": "tr_123",
            "orgId": "org_1",
            "mcpServer": "filesystem",
            "operation": "tools/call",
            "status": "success",
            "startTime": "2026-01-01T00:00:00Z",
        }

        result = client.traces.get("tr_123")

        mock_request.assert_called_once_with("GET", "/v1/traces/tr_123")


class TestCostsClient:
    """Tests for costs client."""

    def test_costs_client_exists(self, client):
        """Client should provide costs client."""
        costs = client.costs
        assert costs is not None

    @patch.object(GatewayOps, "_request")
    def test_costs_summary(self, mock_request, client):
        """costs.summary() should call correct endpoint."""
        mock_request.return_value = {
            "total_cost": 100.0,
            "total_requests": 1000,
            "avg_cost_per_request": 0.1,
            "period": "month",
            "start_date": "2025-12-01T00:00:00Z",
            "end_date": "2026-01-01T00:00:00Z",
        }

        result = client.costs.summary(period="month")

        mock_request.assert_called_once()
        assert result.total_cost == 100.0


class TestErrorHandling:
    """Tests for error handling."""

    def test_authentication_error(self, client, mock_response):
        """Client should raise AuthenticationError for 401."""
        response = mock_response(401, {
            "error": {"code": "unauthorized", "message": "Invalid API key"}
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(AuthenticationError) as exc_info:
                client._request("GET", "/test")
            assert "Invalid API key" in str(exc_info.value)

    def test_rate_limit_error(self, client, mock_response):
        """Client should raise RateLimitError for 429."""
        response = mock_response(429, {
            "error": {"code": "rate_limit_exceeded", "message": "Too many requests"}
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(RateLimitError):
                client._request("GET", "/test")

    def test_not_found_error(self, client, mock_response):
        """Client should raise NotFoundError for 404."""
        response = mock_response(404, {
            "error": {"code": "not_found", "message": "Resource not found"}
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(NotFoundError):
                client._request("GET", "/test")

    def test_validation_error(self, client, mock_response):
        """Client should raise ValidationError for 400."""
        response = mock_response(400, {
            "error": {"code": "validation_error", "message": "Invalid input"}
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(ValidationError):
                client._request("POST", "/test")

    def test_injection_detected_error(self, client, mock_response):
        """Client should raise InjectionDetectedError for injection."""
        response = mock_response(400, {
            "error": {
                "code": "injection_detected",
                "message": "Prompt injection detected",
                "details": {"pattern": "ignore instructions", "severity": "high"}
            }
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(InjectionDetectedError) as exc_info:
                client._request("POST", "/test")
            assert exc_info.value.severity == "high"

    def test_tool_access_denied_error(self, client, mock_response):
        """Client should raise ToolAccessDeniedError for 403 tool access."""
        response = mock_response(403, {
            "error": {
                "code": "tool_access_denied",
                "message": "Tool requires approval",
                "details": {
                    "mcp_server": "filesystem",
                    "tool_name": "delete_file",
                    "requires_approval": True
                }
            }
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(ToolAccessDeniedError) as exc_info:
                client._request("POST", "/test")
            assert exc_info.value.requires_approval is True

    def test_server_error(self, client, mock_response):
        """Client should raise ServerError for 500."""
        response = mock_response(500, {
            "error": {"code": "internal_error", "message": "Server error"}
        })

        with patch.object(client._client, "request", return_value=response):
            with pytest.raises(ServerError):
                client._request("GET", "/test")

    def test_network_error_on_timeout(self, client):
        """Client should raise NetworkError on timeout."""
        with patch.object(client._client, "request", side_effect=httpx.TimeoutException("timeout")):
            with pytest.raises(NetworkError):
                client._request("GET", "/test")

    def test_network_error_on_connection_error(self, client):
        """Client should raise NetworkError on connection error."""
        with patch.object(client._client, "request", side_effect=httpx.RequestError("connection failed")):
            with pytest.raises(NetworkError):
                client._request("GET", "/test")


class TestTraceContext:
    """Tests for trace context."""

    def test_trace_context_sets_id(self, client):
        """trace() should create context with ID."""
        with client.trace("test-operation") as ctx:
            assert ctx.trace_id is not None
            assert ctx.name == "test-operation"
            assert client._trace_context == ctx.trace_id

    def test_trace_context_clears_after(self, client):
        """trace() should clear context after exiting."""
        with client.trace("test-operation"):
            pass
        assert client._trace_context is None
