'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, Copy, Trash2, Key, MoreVertical, Loader2, AlertTriangle, RotateCw, CheckCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useApiKeys, useCreateApiKey, useDeleteApiKey, useRotateApiKey } from '@/lib/hooks/use-api';
import { mutate } from 'swr';

export default function APIKeysPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyEnv, setNewKeyEnv] = useState('development');
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const { data, isLoading, error } = useApiKeys();
  const createMutation = useCreateApiKey();
  const deleteMutation = useDeleteApiKey();
  const rotateMutation = useRotateApiKey();

  const handleCreate = async () => {
    if (!newKeyName.trim()) return;

    try {
      const result = await createMutation.trigger({
        name: newKeyName,
        environment: newKeyEnv,
      });
      setCreatedKey(result.raw_key);
      setNewKeyName('');
      mutate('api-keys');
    } catch (err) {
      console.error('Failed to create key:', err);
    }
  };

  const handleDelete = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
      return;
    }

    try {
      await deleteMutation.trigger(keyId);
      mutate('api-keys');
    } catch (err) {
      console.error('Failed to delete key:', err);
    }
  };

  const handleRotate = async (keyId: string) => {
    if (!confirm('This will revoke the current key and generate a new one. Continue?')) {
      return;
    }

    try {
      const result = await rotateMutation.trigger(keyId);
      setCreatedKey(result.raw_key);
      mutate('api-keys');
    } catch (err) {
      console.error('Failed to rotate key:', err);
    }
  };

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-gray-600">Failed to load API keys</p>
        </div>
      </div>
    );
  }

  const apiKeys = data?.api_keys || [];

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

      {/* Create Modal */}
      {showCreateModal && (
        <Card className="border-indigo-200 bg-indigo-50/50">
          <CardHeader>
            <CardTitle className="text-lg font-medium">Create New API Key</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="e.g., Production API"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Environment</label>
              <select
                value={newKeyEnv}
                onChange={(e) => setNewKeyEnv(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                <option value="development">Development</option>
                <option value="staging">Staging</option>
                <option value="production">Production</option>
              </select>
            </div>
            <div className="flex gap-2">
              <Button onClick={handleCreate} disabled={createMutation.isMutating}>
                {createMutation.isMutating ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Plus className="h-4 w-4 mr-2" />
                )}
                Create Key
              </Button>
              <Button variant="outline" onClick={() => {
                setShowCreateModal(false);
                setCreatedKey(null);
              }}>
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Show created key */}
      {createdKey && (
        <Card className="border-green-200 bg-green-50">
          <CardContent className="py-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-green-800 mb-1">API Key Created Successfully!</p>
                <p className="text-xs text-green-600 mb-2">Copy this key now - you won't be able to see it again.</p>
                <code className="text-sm text-green-800 font-mono bg-green-100 px-2 py-1 rounded">{createdKey}</code>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleCopy(createdKey, 'new')}
                className="border-green-300 text-green-700 hover:bg-green-100"
              >
                {copiedId === 'new' ? (
                  <CheckCircle className="h-4 w-4 mr-1" />
                ) : (
                  <Copy className="h-4 w-4 mr-1" />
                )}
                {copiedId === 'new' ? 'Copied!' : 'Copy'}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

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
                  <tr key={key.id} className={cn("hover:bg-gray-50", key.revoked && "opacity-50")}>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="rounded-full bg-gray-100 p-2">
                          <Key className="h-4 w-4 text-gray-600" />
                        </div>
                        <div>
                          <span className="font-medium text-gray-900">{key.name}</span>
                          {key.revoked && (
                            <span className="ml-2 text-xs text-red-600">(Revoked)</span>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <code className="text-sm text-gray-600 font-mono">{key.key_prefix}...</code>
                        <button
                          className="text-gray-400 hover:text-gray-600"
                          onClick={() => handleCopy(key.key_prefix + '...', key.id)}
                        >
                          {copiedId === key.id ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={cn(
                          'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
                          key.environment === 'production'
                            ? 'bg-green-100 text-green-800'
                            : key.environment === 'staging'
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-blue-100 text-blue-800'
                        )}
                      >
                        {key.environment}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {key.permissions?.join(', ') || 'full'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">{key.rate_limit} rpm</td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-6 py-4">
                      {!key.revoked && (
                        <div className="flex items-center gap-2">
                          <button
                            className="text-gray-400 hover:text-indigo-600"
                            onClick={() => handleRotate(key.id)}
                            title="Rotate key"
                          >
                            <RotateCw className="h-4 w-4" />
                          </button>
                          <button
                            className="text-gray-400 hover:text-red-600"
                            onClick={() => handleDelete(key.id)}
                            title="Revoke key"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
                {apiKeys.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center text-gray-500">
                      No API keys found. Create your first key to get started.
                    </td>
                  </tr>
                )}
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
