"""Tests for GatewayOps SDK types."""

import pytest
from datetime import datetime
from gatewayops.types import (
    TracePage,
    Trace,
    CostSummary,
    ToolDefinition,
    ToolCallResult,
    Resource,
    Span,
)


class TestTracePage:
    """Tests for TracePage model."""

    def test_empty_traces_from_null(self):
        """TracePage should handle null traces array."""
        data = {
            "traces": None,
            "total": 0,
            "limit": 20,
            "offset": 0,
        }
        page = TracePage(**data)
        assert page.traces == []
        assert page.total == 0
        assert page.has_more is False

    def test_empty_traces_from_empty_list(self):
        """TracePage should handle empty traces array."""
        data = {
            "traces": [],
            "total": 0,
            "limit": 20,
            "offset": 0,
        }
        page = TracePage(**data)
        assert page.traces == []
        assert page.has_more is False

    def test_has_more_true(self):
        """has_more should be True when more traces exist."""
        data = {
            "traces": [],
            "total": 100,
            "limit": 20,
            "offset": 0,
        }
        page = TracePage(**data)
        assert page.has_more is True

    def test_has_more_false_at_end(self):
        """has_more should be False at end of results."""
        data = {
            "traces": [],
            "total": 100,
            "limit": 20,
            "offset": 80,
        }
        page = TracePage(**data)
        assert page.has_more is False

    def test_has_more_false_exact(self):
        """has_more should be False when exactly at total."""
        data = {
            "traces": [],
            "total": 20,
            "limit": 20,
            "offset": 0,
        }
        page = TracePage(**data)
        assert page.has_more is False


class TestCostSummary:
    """Tests for CostSummary model."""

    def test_snake_case_fields(self):
        """CostSummary should use snake_case field names from API."""
        data = {
            "total_cost": 123.45,
            "total_requests": 1000,
            "avg_cost_per_request": 0.12345,
            "period": "month",
            "start_date": "2025-12-01T00:00:00Z",
            "end_date": "2026-01-01T00:00:00Z",
        }
        summary = CostSummary(**data)
        assert summary.total_cost == 123.45
        assert summary.total_requests == 1000
        assert summary.avg_cost_per_request == 0.12345
        assert summary.period == "month"

    def test_backwards_compatible_aliases(self):
        """CostSummary should provide backwards-compatible property aliases."""
        data = {
            "total_cost": 100.0,
            "total_requests": 500,
            "avg_cost_per_request": 0.2,
            "period": "week",
            "start_date": "2025-12-25T00:00:00Z",
            "end_date": "2026-01-01T00:00:00Z",
        }
        summary = CostSummary(**data)
        # Test aliases
        assert summary.period_start == summary.start_date
        assert summary.period_end == summary.end_date
        assert summary.request_count == summary.total_requests

    def test_default_values(self):
        """CostSummary should have sensible defaults."""
        summary = CostSummary()
        assert summary.total_cost == 0.0
        assert summary.total_requests == 0
        assert summary.period == "month"


class TestToolDefinition:
    """Tests for ToolDefinition model."""

    def test_basic_tool(self):
        """ToolDefinition should parse basic tool data."""
        data = {
            "name": "read_file",
            "description": "Read a file from the filesystem",
        }
        tool = ToolDefinition(**data)
        assert tool.name == "read_file"
        assert tool.description == "Read a file from the filesystem"

    def test_tool_with_schema(self):
        """ToolDefinition should parse tool with input schema."""
        data = {
            "name": "write_file",
            "description": "Write content to a file",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "path": {"type": "string"},
                    "content": {"type": "string"},
                },
                "required": ["path", "content"],
            },
        }
        tool = ToolDefinition(**data)
        assert tool.name == "write_file"
        assert tool.input_schema is not None
        assert tool.input_schema["type"] == "object"


class TestToolCallResult:
    """Tests for ToolCallResult model."""

    def test_successful_result(self):
        """ToolCallResult should parse successful result."""
        data = {
            "content": {"data": "file contents here"},
            "isError": False,
            "traceId": "tr_abc123",
            "durationMs": 45,
        }
        result = ToolCallResult(**data)
        assert result.content == {"data": "file contents here"}
        assert result.is_error is False
        assert result.trace_id == "tr_abc123"
        assert result.duration_ms == 45

    def test_error_result(self):
        """ToolCallResult should parse error result."""
        data = {
            "content": "File not found",
            "isError": True,
        }
        result = ToolCallResult(**data)
        assert result.is_error is True


class TestResource:
    """Tests for Resource model."""

    def test_basic_resource(self):
        """Resource should parse basic resource data."""
        data = {
            "uri": "file:///data/report.csv",
            "name": "report.csv",
            "description": "Monthly report",
            "mimeType": "text/csv",
        }
        resource = Resource(**data)
        assert resource.uri == "file:///data/report.csv"
        assert resource.name == "report.csv"
        assert resource.mime_type == "text/csv"


class TestSpan:
    """Tests for Span model."""

    def test_basic_span(self):
        """Span should parse basic span data."""
        data = {
            "id": "span_123",
            "traceId": "tr_abc",
            "name": "authenticate",
            "kind": "internal",
            "status": "success",
            "startTime": "2026-01-01T00:00:00Z",
            "durationMs": 5,
        }
        span = Span(**data)
        assert span.id == "span_123"
        assert span.trace_id == "tr_abc"
        assert span.name == "authenticate"
        assert span.duration_ms == 5

    def test_span_with_parent(self):
        """Span should parse span with parent."""
        data = {
            "id": "span_456",
            "traceId": "tr_abc",
            "parentSpanId": "span_123",
            "name": "validate_request",
            "kind": "internal",
            "status": "success",
            "startTime": "2026-01-01T00:00:00Z",
        }
        span = Span(**data)
        assert span.parent_span_id == "span_123"
