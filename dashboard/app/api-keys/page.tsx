'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, Copy, Trash2, Key, MoreVertical } from 'lucide-react';
import { cn } from '@/lib/utils';

// Sample data
const apiKeys = [
  {
    id: 'key_1',
    name: 'Production API',
    prefix: 'gwo_prd_abc1',
    environment: 'production',
    permissions: 'full',
    rateLimitRpm: 1000,
    createdAt: '2024-01-10T10:00:00Z',
    lastUsedAt: '2024-01-15T14:30:00Z',
  },
  {
    id: 'key_2',
    name: 'Development',
    prefix: 'gwo_dev_xyz2',
    environment: 'sandbox',
    permissions: 'full',
    rateLimitRpm: 100,
    createdAt: '2024-01-08T10:00:00Z',
    lastUsedAt: '2024-01-15T12:00:00Z',
  },
  {
    id: 'key_3',
    name: 'CI/CD Pipeline',
    prefix: 'gwo_prd_def3',
    environment: 'production',
    permissions: 'read',
    rateLimitRpm: 500,
    createdAt: '2024-01-05T10:00:00Z',
    lastUsedAt: '2024-01-14T08:00:00Z',
  },
];

export default function APIKeysPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">API Keys</h1>
          <p className="text-gray-500">Manage your GatewayOps API keys</p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create API Key
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Name</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Key</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Environment</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Permissions</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Rate Limit</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500">Last Used</th>
                  <th className="px-6 py-4 text-left text-sm font-medium text-gray-500"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {apiKeys.map((key) => (
                  <tr key={key.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="rounded-full bg-gray-100 p-2">
                          <Key className="h-4 w-4 text-gray-600" />
                        </div>
                        <span className="font-medium text-gray-900">{key.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <code className="text-sm text-gray-600 font-mono">{key.prefix}...</code>
                        <button className="text-gray-400 hover:text-gray-600">
                          <Copy className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={cn(
                          'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
                          key.environment === 'production'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-yellow-100 text-yellow-800'
                        )}
                      >
                        {key.environment}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">{key.permissions}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{key.rateLimitRpm} rpm</td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {new Date(key.lastUsedAt).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <button className="text-gray-400 hover:text-red-600">
                          <Trash2 className="h-4 w-4" />
                        </button>
                        <button className="text-gray-400 hover:text-gray-600">
                          <MoreVertical className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Usage Guidelines</CardTitle>
        </CardHeader>
        <CardContent className="prose prose-sm text-gray-600">
          <ul className="space-y-2">
            <li>Production keys should only be used in production environments.</li>
            <li>Rotate keys regularly for security best practices.</li>
            <li>Use separate keys for different applications or services.</li>
            <li>Set appropriate rate limits to prevent abuse.</li>
            <li>Never expose API keys in client-side code or version control.</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
