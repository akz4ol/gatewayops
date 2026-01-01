'use client';

import Link from 'next/link';
import { CheckCircle, XCircle, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';

// Sample data - in production this would come from the API
const traces = [
  {
    id: 'tr_1a2b3c4d',
    server: 'filesystem',
    operation: 'tools/call',
    tool: 'read_file',
    status: 'success',
    duration: 45,
    time: '2 min ago',
  },
  {
    id: 'tr_2b3c4d5e',
    server: 'database',
    operation: 'tools/call',
    tool: 'query',
    status: 'success',
    duration: 128,
    time: '5 min ago',
  },
  {
    id: 'tr_3c4d5e6f',
    server: 'github',
    operation: 'tools/call',
    tool: 'create_issue',
    status: 'error',
    duration: 2340,
    time: '8 min ago',
  },
  {
    id: 'tr_4d5e6f7g',
    server: 'slack',
    operation: 'tools/call',
    tool: 'send_message',
    status: 'success',
    duration: 89,
    time: '12 min ago',
  },
  {
    id: 'tr_5e6f7g8h',
    server: 'filesystem',
    operation: 'resources/read',
    tool: 'file://config.json',
    status: 'success',
    duration: 23,
    time: '15 min ago',
  },
];

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

export function RecentTraces() {
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
            const config = statusConfig[trace.status as keyof typeof statusConfig];
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
                <td className="py-3 text-sm text-gray-500">{trace.time}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
