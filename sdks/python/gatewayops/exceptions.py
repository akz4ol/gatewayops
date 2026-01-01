"""GatewayOps SDK exceptions."""

from typing import Any, Dict, Optional


class GatewayOpsError(Exception):
    """Base exception for GatewayOps SDK errors."""

    def __init__(
        self,
        message: str,
        code: Optional[str] = None,
        status_code: Optional[int] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message)
        self.message = message
        self.code = code
        self.status_code = status_code
        self.details = details or {}

    def __str__(self) -> str:
        if self.code:
            return f"[{self.code}] {self.message}"
        return self.message


class AuthenticationError(GatewayOpsError):
    """Raised when authentication fails."""

    def __init__(
        self,
        message: str = "Authentication failed",
        code: str = "unauthorized",
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code=code, status_code=401, details=details)


class RateLimitError(GatewayOpsError):
    """Raised when rate limit is exceeded."""

    def __init__(
        self,
        message: str = "Rate limit exceeded",
        retry_after: Optional[int] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="rate_limit_exceeded", status_code=429, details=details)
        self.retry_after = retry_after


class NotFoundError(GatewayOpsError):
    """Raised when a resource is not found."""

    def __init__(
        self,
        message: str = "Resource not found",
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="not_found", status_code=404, details=details)
        self.resource_type = resource_type
        self.resource_id = resource_id


class ValidationError(GatewayOpsError):
    """Raised when request validation fails."""

    def __init__(
        self,
        message: str = "Validation error",
        field: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="validation_error", status_code=400, details=details)
        self.field = field


class InjectionDetectedError(GatewayOpsError):
    """Raised when prompt injection is detected."""

    def __init__(
        self,
        message: str = "Potential prompt injection detected",
        pattern: Optional[str] = None,
        severity: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="injection_detected", status_code=400, details=details)
        self.pattern = pattern
        self.severity = severity


class ToolAccessDeniedError(GatewayOpsError):
    """Raised when tool access is denied."""

    def __init__(
        self,
        message: str = "Tool access denied",
        mcp_server: Optional[str] = None,
        tool_name: Optional[str] = None,
        requires_approval: bool = False,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="tool_access_denied", status_code=403, details=details)
        self.mcp_server = mcp_server
        self.tool_name = tool_name
        self.requires_approval = requires_approval


class ServerError(GatewayOpsError):
    """Raised when the server returns an error."""

    def __init__(
        self,
        message: str = "Server error",
        code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code=code or "server_error", status_code=500, details=details)


class TimeoutError(GatewayOpsError):
    """Raised when a request times out."""

    def __init__(
        self,
        message: str = "Request timed out",
        timeout_seconds: Optional[float] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="timeout", status_code=408, details=details)
        self.timeout_seconds = timeout_seconds


class NetworkError(GatewayOpsError):
    """Raised when a network error occurs."""

    def __init__(
        self,
        message: str = "Network error",
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message, code="network_error", details=details)
