'use client';

import { Loader2 } from 'lucide-react';
import { useTopServers } from '@/lib/hooks/use-api';

export function TopServers() {
  const { data, isLoading, error } = useTopServers();

  if (isLoading) {
    return (
      <div className="h-[200px] flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
      </div>
    );
  }

  if (error || !data?.servers) {
    return (
      <div className="h-[200px] flex items-center justify-center text-gray-500">
        Failed to load server data
      </div>
    );
  }

  const servers = data.servers;
  const maxRequests = Math.max(...servers.map((s) => s.requests));

  return (
    <div className="space-y-4">
      {servers.map((server) => (
        <div key={server.name} className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium text-gray-900">{server.name}</span>
            <span className="text-gray-500">
              {(server.requests / 1000).toFixed(1)}K requests
            </span>
          </div>
          <div className="relative h-2 overflow-hidden rounded-full bg-gray-100">
            <div
              className="absolute inset-y-0 left-0 bg-indigo-500 rounded-full"
              style={{ width: `${(server.requests / maxRequests) * 100}%` }}
            />
          </div>
          <div className="flex justify-end">
            <span className="text-xs text-gray-500">${server.cost.toFixed(2)}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
