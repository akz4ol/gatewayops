'use client';

import Link from 'next/link';
import { CheckCircle, XCircle, Clock, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useRecentTraces } from '@/lib/hooks/use-api';

const statusConfig = {
  success: {
    icon: CheckCircle,
    color: 'text-green-600',
    bg: 'bg-green-50',
  },
  error: {
    icon: XCircle,
    color: 'text-red-600',
    bg: 'bg-red-50',
  },
  timeout: {
    icon: Clock,
    color: 'text-yellow-600',
    bg: 'bg-yellow-50',
  },
};

function formatTime(isoTime: string): string {
  const date = new Date(isoTime);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} min ago`;

  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

export function RecentTraces() {
  const { data, isLoading, error } = useRecentTraces();

  if (isLoading) {
    return (
      <div className="h-[200px] flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
      </div>
    );
  }

  if (error || !data?.traces) {
    return (
      <div className="h-[200px] flex items-center justify-center text-gray-500">
        Failed to load traces
      </div>
    );
  }

  const traces = data.traces;

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-gray-200">
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Status</th>
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Trace ID</th>
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Server</th>
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Operation</th>
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Duration</th>
            <th className="pb-3 text-left text-sm font-medium text-gray-500">Time</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {traces.map((trace) => {
            const config = statusConfig[trace.status as keyof typeof statusConfig] || statusConfig.success;
            const Icon = config.icon;
            return (
              <tr key={trace.id} className="hover:bg-gray-50">
                <td className="py-3">
                  <div className={cn('inline-flex rounded-full p-1', config.bg)}>
                    <Icon className={cn('h-4 w-4', config.color)} />
                  </div>
                </td>
                <td className="py-3">
                  <Link
                    href={`/traces/${trace.id}`}
                    className="font-mono text-sm text-indigo-600 hover:underline"
                  >
                    {trace.id}
                  </Link>
                </td>
                <td className="py-3">
                  <span className="rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-700">
                    {trace.server}
                  </span>
                </td>
                <td className="py-3 text-sm text-gray-900">{trace.tool}</td>
                <td className="py-3 text-sm text-gray-500">{trace.duration}ms</td>
                <td className="py-3 text-sm text-gray-500">{formatTime(trace.time)}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
