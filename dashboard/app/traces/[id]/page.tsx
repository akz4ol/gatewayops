'use client';

import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ArrowLeft, CheckCircle, XCircle, Clock, Copy } from 'lucide-react';
import { cn } from '@/lib/utils';
import { TraceWaterfall } from '@/components/traces/waterfall';

// Sample trace data
const trace = {
  id: 'tr_1a2b3c4d5e6f',
  server: 'filesystem',
  operation: 'tools/call',
  tool: 'read_file',
  status: 'success',
  startTime: '2024-01-15T10:30:00Z',
  endTime: '2024-01-15T10:30:00.145Z',
  duration: 145,
  cost: 0.0012,
  apiKeyId: 'key_prod_abc123',
  spans: [
    {
      id: 'span_1',
      name: 'gateway.receive',
      startTime: 0,
      duration: 5,
      status: 'ok',
    },
    {
      id: 'span_2',
      name: 'auth.validate',
      startTime: 5,
      duration: 12,
      status: 'ok',
    },
    {
      id: 'span_3',
      name: 'rbac.check',
      startTime: 17,
      duration: 8,
      status: 'ok',
    },
    {
      id: 'span_4',
      name: 'injection.scan',
      startTime: 25,
      duration: 15,
      status: 'ok',
    },
    {
      id: 'span_5',
      name: 'mcp.connect',
      startTime: 40,
      duration: 25,
      status: 'ok',
    },
    {
      id: 'span_6',
      name: 'mcp.tools/call',
      startTime: 65,
      duration: 70,
      status: 'ok',
    },
    {
      id: 'span_7',
      name: 'gateway.respond',
      startTime: 135,
      duration: 10,
      status: 'ok',
    },
  ],
  request: {
    tool: 'read_file',
    arguments: {
      path: '/data/config.json',
    },
  },
  response: {
    content: '{"database": "postgresql://...", "redis": "redis://..."}',
    isError: false,
  },
};

export default function TraceDetailPage({ params }: { params: { id: string } }) {
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
            <div className="flex items-center gap-1.5 rounded-full bg-green-50 px-2 py-1">
              <CheckCircle className="h-3.5 w-3.5 text-green-600" />
              <span className="text-xs font-medium text-green-600">Success</span>
            </div>
          </div>
          <div className="flex items-center gap-2 mt-1">
            <code className="text-sm text-gray-500 font-mono">{params.id}</code>
            <button className="text-gray-400 hover:text-gray-600">
              <Copy className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Server</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="rounded-md bg-gray-100 px-2 py-1 text-sm font-medium text-gray-700">
              {trace.server}
            </span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Duration</CardTitle>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold text-gray-900">{trace.duration}ms</span>
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

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Trace Waterfall</CardTitle>
        </CardHeader>
        <CardContent>
          <TraceWaterfall spans={trace.spans} totalDuration={trace.duration} />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Request</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded-lg bg-gray-900 p-4 text-sm text-gray-100 overflow-x-auto">
              {JSON.stringify(trace.request, null, 2)}
            </pre>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Response</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded-lg bg-gray-900 p-4 text-sm text-gray-100 overflow-x-auto">
              {JSON.stringify(trace.response, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
