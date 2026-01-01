'use client';

import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ArrowLeft, CheckCircle, XCircle, Clock, Copy, Loader2, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { TraceWaterfall } from '@/components/traces/waterfall';
import { useTrace } from '@/lib/hooks/use-api';

const statusConfig = {
  success: { icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-50', label: 'Success' },
  error: { icon: XCircle, color: 'text-red-600', bg: 'bg-red-50', label: 'Error' },
  timeout: { icon: Clock, color: 'text-yellow-600', bg: 'bg-yellow-50', label: 'Timeout' },
};

export default function TraceDetailPage({ params }: { params: { id: string } }) {
  const { data, isLoading, error } = useTrace(params.id);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-gray-600">Failed to load trace details</p>
          <Link href="/traces" className="text-indigo-600 hover:underline text-sm mt-2 block">
            Back to traces
          </Link>
        </div>
      </div>
    );
  }

  const { trace, spans } = data;
  const config = statusConfig[trace.status as keyof typeof statusConfig] || statusConfig.success;
  const StatusIcon = config.icon;

  // Convert spans to waterfall format
  const waterfallSpans = spans.map((span, index) => {
    const baseTime = new Date(spans[0]?.start_time || trace.created_at).getTime();
    const spanStartTime = new Date(span.start_time).getTime();
    return {
      id: span.span_id,
      name: span.name,
      startTime: spanStartTime - baseTime,
      duration: span.duration_ms,
      status: span.status === 'success' ? 'ok' : 'error',
    };
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/traces">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-5 w-5" />
          </Button>
        </Link>
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-gray-900">Trace Details</h1>
            <div className={cn('flex items-center gap-1.5 rounded-full px-2 py-1', config.bg)}>
              <StatusIcon className={cn('h-3.5 w-3.5', config.color)} />
              <span className={cn('text-xs font-medium', config.color)}>{config.label}</span>
            </div>
          </div>
          <div className="flex items-center gap-2 mt-1">
            <code className="text-sm text-gray-500 font-mono">{trace.trace_id}</code>
            <button
              className="text-gray-400 hover:text-gray-600"
              onClick={() => navigator.clipboard.writeText(trace.trace_id)}
            >
              <Copy className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Server</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="rounded-md bg-gray-100 px-2 py-1 text-sm font-medium text-gray-700">
              {trace.mcp_server}
            </span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Operation</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="text-sm font-medium text-gray-900">{trace.operation}</span>
            <p className="text-xs text-gray-500 mt-1">{trace.tool_name}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Duration</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold text-gray-900">{trace.duration_ms}ms</span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold text-gray-900">${trace.cost.toFixed(4)}</span>
          </CardContent>
        </Card>
      </div>

      {trace.error_msg && (
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-lg font-medium text-red-800">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-red-700">{trace.error_msg}</p>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Trace Waterfall</CardTitle>
        </CardHeader>
        <CardContent>
          <TraceWaterfall spans={waterfallSpans} totalDuration={trace.duration_ms} />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Request Info</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded-lg bg-gray-900 p-4 text-sm text-gray-100 overflow-x-auto">
{JSON.stringify({
  server: trace.mcp_server,
  operation: trace.operation,
  tool: trace.tool_name,
  request_size: trace.request_size,
}, null, 2)}
            </pre>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Response Info</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded-lg bg-gray-900 p-4 text-sm text-gray-100 overflow-x-auto">
{JSON.stringify({
  status_code: trace.status_code,
  response_size: trace.response_size,
  duration_ms: trace.duration_ms,
  cost: trace.cost,
}, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Spans ({spans.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {spans.map((span) => (
              <div key={span.span_id} className="flex items-center justify-between py-2 px-3 bg-gray-50 rounded">
                <div>
                  <span className="font-medium text-gray-900">{span.name}</span>
                  <span className="text-gray-500 text-sm ml-2">({span.kind})</span>
                </div>
                <div className="flex items-center gap-4 text-sm">
                  <span className={span.status === 'success' ? 'text-green-600' : 'text-red-600'}>
                    {span.status}
                  </span>
                  <span className="text-gray-500">{span.duration_ms}ms</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
