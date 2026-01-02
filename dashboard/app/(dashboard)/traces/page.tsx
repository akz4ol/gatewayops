'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { CheckCircle, XCircle, Clock, Search, Loader2, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTraces } from '@/lib/hooks/use-api';

const statusConfig = {
  success: { icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-50', label: 'Success' },
  error: { icon: XCircle, color: 'text-red-600', bg: 'bg-red-50', label: 'Error' },
  timeout: { icon: Clock, color: 'text-yellow-600', bg: 'bg-yellow-50', label: 'Timeout' },
};

export default function TracesPage() {
  const [filter, setFilter] = useState({ server: '', status: '' });
  const { data, isLoading, error } = useTraces({
    limit: 20,
    server: filter.server || undefined,
    status: filter.status || undefined,
  });

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-gray-600">Failed to load traces</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Traces</h1>
          <p className="text-gray-500">View and analyze MCP request traces</p>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search by trace ID, server, or tool..."
                className="h-10 w-full rounded-md border border-gray-200 bg-white pl-10 pr-4 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
            <select
              value={filter.server}
              onChange={(e) => setFilter({ ...filter, server: e.target.value })}
              className="h-10 rounded-md border border-gray-200 bg-white px-3 text-sm"
            >
              <option value="">All Servers</option>
              <option value="filesystem">filesystem</option>
              <option value="database">database</option>
              <option value="github">github</option>
              <option value="slack">slack</option>
            </select>
            <select
              value={filter.status}
              onChange={(e) => setFilter({ ...filter, status: e.target.value })}
              className="h-10 rounded-md border border-gray-200 bg-white px-3 text-sm"
            >
              <option value="">All Status</option>
              <option value="success">Success</option>
              <option value="error">Error</option>
              <option value="timeout">Timeout</option>
            </select>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Status</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Trace ID</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Server</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Operation</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Tool/Resource</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Duration</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Cost</th>
                    <th className="pb-3 text-left text-sm font-medium text-gray-500">Time</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {data?.traces?.map((trace) => {
                    const config = statusConfig[trace.status as keyof typeof statusConfig] || statusConfig.success;
                    const Icon = config.icon;
                    return (
                      <tr key={trace.id} className="hover:bg-gray-50">
                        <td className="py-3">
                          <div className={cn('inline-flex items-center gap-1.5 rounded-full px-2 py-1', config.bg)}>
                            <Icon className={cn('h-3.5 w-3.5', config.color)} />
                            <span className={cn('text-xs font-medium', config.color)}>{config.label}</span>
                          </div>
                        </td>
                        <td className="py-3">
                          <Link
                            href={`/traces/${trace.trace_id}`}
                            className="font-mono text-sm text-indigo-600 hover:underline"
                          >
                            {trace.trace_id}
                          </Link>
                        </td>
                        <td className="py-3">
                          <span className="rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-700">
                            {trace.mcp_server}
                          </span>
                        </td>
                        <td className="py-3 text-sm text-gray-500">{trace.operation}</td>
                        <td className="py-3 text-sm text-gray-900">{trace.tool_name}</td>
                        <td className="py-3 text-sm text-gray-500">{trace.duration_ms}ms</td>
                        <td className="py-3 text-sm text-gray-500">${trace.cost.toFixed(4)}</td>
                        <td className="py-3 text-sm text-gray-500">
                          {new Date(trace.created_at).toLocaleTimeString()}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {data && (
                <div className="mt-4 flex items-center justify-between text-sm text-gray-500">
                  <span>Showing {data.traces?.length || 0} of {data.total} traces</span>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
