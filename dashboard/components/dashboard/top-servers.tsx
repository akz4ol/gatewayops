'use client';

// Sample data - in production this would come from the API
const servers = [
  { name: 'filesystem', requests: 45000, cost: 1250.0 },
  { name: 'database', requests: 32000, cost: 890.0 },
  { name: 'github', requests: 28000, cost: 780.0 },
  { name: 'slack', requests: 15000, cost: 420.0 },
  { name: 'memory', requests: 12000, cost: 340.0 },
];

const maxRequests = Math.max(...servers.map((s) => s.requests));

export function TopServers() {
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
