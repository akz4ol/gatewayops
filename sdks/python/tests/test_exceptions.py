"""Tests for GatewayOps SDK exceptions."""

import pytest
from gatewayops.exceptions import (
    GatewayOpsError,
    AuthenticationError,
    RateLimitError,
    NotFoundError,
    ValidationError,
    InjectionDetectedError,
    ToolAccessDeniedError,
    ServerError,
    TimeoutError,
    NetworkError,
)


class TestGatewayOpsError:
    """Tests for base GatewayOpsError."""

    def test_basic_error(self):
        """GatewayOpsError should store message."""
        error = GatewayOpsError("Something went wrong")
        assert str(error) == "Something went wrong"
        assert error.message == "Something went wrong"

    def test_error_with_code(self):
        """GatewayOpsError should format with code."""
        error = GatewayOpsError("Failed", code="some_error")
        assert str(error) == "[some_error] Failed"
        assert error.code == "some_error"

    def test_error_with_status_code(self):
        """GatewayOpsError should store status code."""
        error = GatewayOpsError("Failed", status_code=500)
        assert error.status_code == 500

    def test_error_with_details(self):
        """GatewayOpsError should store details."""
        error = GatewayOpsError("Failed", details={"key": "value"})
        assert error.details == {"key": "value"}


class TestAuthenticationError:
    """Tests for AuthenticationError."""

    def test_default_message(self):
        """AuthenticationError should have default message."""
        error = AuthenticationError()
        assert "Authentication failed" in str(error)
        assert error.status_code == 401

    def test_custom_message(self):
        """AuthenticationError should accept custom message."""
        error = AuthenticationError("Invalid token")
        assert "Invalid token" in str(error)


class TestRateLimitError:
    """Tests for RateLimitError."""

    def test_default_message(self):
        """RateLimitError should have default message."""
        error = RateLimitError()
        assert "Rate limit exceeded" in str(error)
        assert error.status_code == 429

    def test_retry_after(self):
        """RateLimitError should store retry_after."""
        error = RateLimitError(retry_after=60)
        assert error.retry_after == 60


class TestNotFoundError:
    """Tests for NotFoundError."""

    def test_default_message(self):
        """NotFoundError should have default message."""
        error = NotFoundError()
        assert "not found" in str(error).lower()
        assert error.status_code == 404

    def test_resource_info(self):
        """NotFoundError should store resource info."""
        error = NotFoundError(
            message="Trace not found",
            resource_type="trace",
            resource_id="tr_123"
        )
        assert error.resource_type == "trace"
        assert error.resource_id == "tr_123"


class TestValidationError:
    """Tests for ValidationError."""

    def test_default_message(self):
        """ValidationError should have default message."""
        error = ValidationError()
        assert error.status_code == 400

    def test_field_info(self):
        """ValidationError should store field info."""
        error = ValidationError(message="Invalid email", field="email")
        assert error.field == "email"


class TestInjectionDetectedError:
    """Tests for InjectionDetectedError."""

    def test_default_message(self):
        """InjectionDetectedError should have default message."""
        error = InjectionDetectedError()
        assert "injection" in str(error).lower()
        assert error.status_code == 400

    def test_pattern_and_severity(self):
        """InjectionDetectedError should store pattern and severity."""
        error = InjectionDetectedError(
            pattern="ignore previous",
            severity="high"
        )
        assert error.pattern == "ignore previous"
        assert error.severity == "high"


class TestToolAccessDeniedError:
    """Tests for ToolAccessDeniedError."""

    def test_default_message(self):
        """ToolAccessDeniedError should have default message."""
        error = ToolAccessDeniedError()
        assert "denied" in str(error).lower()
        assert error.status_code == 403

    def test_tool_info(self):
        """ToolAccessDeniedError should store tool info."""
        error = ToolAccessDeniedError(
            mcp_server="filesystem",
            tool_name="delete_file",
            requires_approval=True
        )
        assert error.mcp_server == "filesystem"
        assert error.tool_name == "delete_file"
        assert error.requires_approval is True


class TestServerError:
    """Tests for ServerError."""

    def test_default_message(self):
        """ServerError should have default message."""
        error = ServerError()
        assert error.status_code == 500

    def test_custom_code(self):
        """ServerError should accept custom code."""
        error = ServerError(code="database_error")
        assert error.code == "database_error"


class TestTimeoutError:
    """Tests for TimeoutError."""

    def test_default_message(self):
        """TimeoutError should have default message."""
        error = TimeoutError()
        assert "timeout" in str(error).lower()
        assert error.status_code == 408

    def test_timeout_seconds(self):
        """TimeoutError should store timeout duration."""
        error = TimeoutError(timeout_seconds=30.0)
        assert error.timeout_seconds == 30.0


class TestNetworkError:
    """Tests for NetworkError."""

    def test_default_message(self):
        """NetworkError should have default message."""
        error = NetworkError()
        assert "network" in str(error).lower()

    def test_no_status_code(self):
        """NetworkError should not have status code."""
        error = NetworkError()
        assert error.status_code is None
