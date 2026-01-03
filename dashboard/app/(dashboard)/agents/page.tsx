'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Bot, Zap, Link2, Globe, Code, Loader2, AlertTriangle, RefreshCw, Copy, Check, Terminal } from 'lucide-react';
import { cn } from '@/lib/utils';

interface AgentStats {
  active: number;
  total: number;
  messages: number;
  by_platform: Record<string, number>;
}

interface OpenAITool {
  type: 'function';
  function: {
    name: string;
    description: string;
    parameters: {
      type: 'object';
      properties: Record<string, { type: string; description: string }>;
      required: string[];
    };
  };
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://gatewayops-api.fly.dev';

export default function AgentsPage() {
  const [stats, setStats] = useState<AgentStats | null>(null);
  const [tools, setTools] = useState<OpenAITool[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState<string | null>(null);

  const fetchData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [statsRes, toolsRes] = await Promise.all([
        fetch(`${API_BASE}/v1/agents/stats`),
        fetch(`${API_BASE}/v1/mcp/tools`),
      ]);

      if (statsRes.ok) {
        setStats(await statsRes.json());
      }
      if (toolsRes.ok) {
        const data = await toolsRes.json();
        setTools(data.tools || []);
      }
    } catch (err) {
      setError('Failed to load agent data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopied(id);
    setTimeout(() => setCopied(null), 2000);
  };

  const pythonExample = `from gatewayops import GatewayOps

gw = GatewayOps(api_key="gwo_prd_...")
conn = gw.agents.connect(platform="my-agent")

# Execute tools
result = conn.call_tool("filesystem", "read_file", {"path": "/data.csv"})`;

  const typescriptExample = `import { GatewayOps } from '@gatewayops/sdk';

const gw = new GatewayOps({ apiKey: 'gwo_prd_...' });
const conn = await gw.agents.connect({ platform: 'my-agent' });

// Execute tools
const result = await conn.callTool('filesystem', 'read_file', { path: '/data.csv' });`;

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <AlertTriangle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-gray-600">{error}</p>
          <button
            onClick={fetchData}
            className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Agent Platform</h1>
          <p className="text-gray-500">Connect AI agents to MCP servers via GatewayOps</p>
        </div>
        <button
          onClick={fetchData}
          disabled={isLoading}
          className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50"
        >
          <RefreshCw className={cn("h-4 w-4", isLoading && "animate-spin")} />
          Refresh
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">Active Connections</p>
                <p className="text-2xl font-bold text-gray-900">
                  {isLoading ? '-' : stats?.active || 0}
                </p>
              </div>
              <div className="h-12 w-12 rounded-lg bg-green-50 flex items-center justify-center">
                <Zap className="h-6 w-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">Total Connections</p>
                <p className="text-2xl font-bold text-gray-900">
                  {isLoading ? '-' : stats?.total || 0}
                </p>
              </div>
              <div className="h-12 w-12 rounded-lg bg-blue-50 flex items-center justify-center">
                <Link2 className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">Messages Sent</p>
                <p className="text-2xl font-bold text-gray-900">
                  {isLoading ? '-' : stats?.messages?.toLocaleString() || 0}
                </p>
              </div>
              <div className="h-12 w-12 rounded-lg bg-purple-50 flex items-center justify-center">
                <Bot className="h-6 w-6 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">Available Tools</p>
                <p className="text-2xl font-bold text-gray-900">
                  {isLoading ? '-' : tools.length}
                </p>
              </div>
              <div className="h-12 w-12 rounded-lg bg-amber-50 flex items-center justify-center">
                <Terminal className="h-6 w-6 text-amber-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Connected Platforms */}
      {stats?.by_platform && Object.keys(stats.by_platform).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Connected Platforms</CardTitle>
            <CardDescription>Active agent connections by platform</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              {Object.entries(stats.by_platform).map(([platform, count]) => (
                <div
                  key={platform}
                  className="flex items-center gap-3 p-4 rounded-lg bg-gray-50"
                >
                  <Globe className="h-5 w-5 text-gray-600" />
                  <div>
                    <p className="font-medium text-gray-900">{platform}</p>
                    <p className="text-sm text-gray-500">{count} connections</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Quick Start */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">Python SDK</CardTitle>
              <button
                onClick={() => copyToClipboard(pythonExample, 'python')}
                className="p-2 text-gray-500 hover:text-gray-700"
              >
                {copied === 'python' ? (
                  <Check className="h-4 w-4 text-green-600" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </button>
            </div>
            <CardDescription>pip install gatewayops</CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg text-sm overflow-x-auto">
              <code>{pythonExample}</code>
            </pre>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">TypeScript SDK</CardTitle>
              <button
                onClick={() => copyToClipboard(typescriptExample, 'typescript')}
                className="p-2 text-gray-500 hover:text-gray-700"
              >
                {copied === 'typescript' ? (
                  <Check className="h-4 w-4 text-green-600" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </button>
            </div>
            <CardDescription>npm install @gatewayops/sdk</CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg text-sm overflow-x-auto">
              <code>{typescriptExample}</code>
            </pre>
          </CardContent>
        </Card>
      </div>

      {/* Available Tools */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Available Tools (OpenAI Format)</CardTitle>
          <CardDescription>
            Tools available for agent integrations in OpenAI-compatible format
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
            </div>
          ) : tools.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No tools available. Connect MCP servers to expose tools.
            </div>
          ) : (
            <div className="space-y-3">
              {tools.map((tool) => (
                <div
                  key={tool.function.name}
                  className="p-4 rounded-lg border border-gray-200 hover:border-gray-300"
                >
                  <div className="flex items-start gap-3">
                    <Code className="h-5 w-5 text-indigo-600 mt-0.5" />
                    <div className="flex-1">
                      <p className="font-mono text-sm font-medium text-gray-900">
                        {tool.function.name}
                      </p>
                      <p className="text-sm text-gray-500 mt-1">
                        {tool.function.description}
                      </p>
                      {tool.function.parameters.required?.length > 0 && (
                        <div className="mt-2 flex gap-2 flex-wrap">
                          {tool.function.parameters.required.map((param) => (
                            <span
                              key={param}
                              className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded-full"
                            >
                              {param}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* API Reference */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">API Endpoints</CardTitle>
          <CardDescription>RESTful API for agent platform integration</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded">POST</span>
                <code className="text-sm font-mono">/v1/agents/connect</code>
              </div>
              <p className="text-sm text-gray-600">Connect an agent platform to the gateway</p>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded">POST</span>
                <code className="text-sm font-mono">/v1/execute</code>
              </div>
              <p className="text-sm text-gray-600">Execute batch tool calls (parallel or sequential)</p>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded">POST</span>
                <code className="text-sm font-mono">/v1/execute/stream</code>
              </div>
              <p className="text-sm text-gray-600">Execute with SSE streaming for real-time progress</p>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded">GET</span>
                <code className="text-sm font-mono">/v1/mcp/tools</code>
              </div>
              <p className="text-sm text-gray-600">List all tools in OpenAI-compatible format</p>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <span className="px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded">GET</span>
                <code className="text-sm font-mono">/v1/agents/stats</code>
              </div>
              <p className="text-sm text-gray-600">Get agent connection statistics</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
