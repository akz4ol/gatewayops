'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';

// Sample data
const costByServer = [
  { name: 'filesystem', cost: 1250, requests: 45000 },
  { name: 'database', cost: 890, requests: 32000 },
  { name: 'github', cost: 780, requests: 28000 },
  { name: 'slack', cost: 420, requests: 15000 },
  { name: 'memory', cost: 340, requests: 12000 },
];

const costByTeam = [
  { name: 'Engineering', value: 2100 },
  { name: 'Data Science', value: 980 },
  { name: 'DevOps', value: 650 },
  { name: 'Product', value: 380 },
  { name: 'Other', value: 120 },
];

const dailyCosts = Array.from({ length: 14 }, (_, i) => ({
  date: new Date(Date.now() - (13 - i) * 86400000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
  cost: Math.random() * 200 + 100,
}));

const COLORS = ['#6366f1', '#3b82f6', '#22c55e', '#eab308', '#ef4444'];

export default function CostsPage() {
  const totalCost = costByServer.reduce((sum, s) => sum + s.cost, 0);
  const totalRequests = costByServer.reduce((sum, s) => sum + s.requests, 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Costs</h1>
          <p className="text-gray-500">Monitor and analyze your MCP Gateway costs</p>
        </div>
        <div className="flex items-center gap-2">
          <select className="h-10 rounded-md border border-gray-200 bg-white px-3 text-sm">
            <option value="month">Last 30 days</option>
            <option value="week">Last 7 days</option>
            <option value="day">Today</option>
          </select>
          <Button variant="outline">Export</Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-gray-900">${totalCost.toFixed(2)}</div>
            <p className="text-sm text-gray-500 mt-1">Last 30 days</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Requests</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-gray-900">{(totalRequests / 1000).toFixed(0)}K</div>
            <p className="text-sm text-gray-500 mt-1">Last 30 days</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Avg Cost/Request</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-gray-900">
              ${(totalCost / totalRequests * 1000).toFixed(2)}
            </div>
            <p className="text-sm text-gray-500 mt-1">Per 1000 requests</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Daily Costs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={dailyCosts}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  tickFormatter={(value) => `$${value}`}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#fff',
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                  }}
                  formatter={(value: number) => [`$${value.toFixed(2)}`, 'Cost']}
                />
                <Bar dataKey="cost" fill="#6366f1" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Cost by Server</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {costByServer.map((server, i) => (
                <div key={server.name} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium text-gray-900">{server.name}</span>
                    <span className="text-gray-500">${server.cost.toFixed(2)}</span>
                  </div>
                  <div className="relative h-2 overflow-hidden rounded-full bg-gray-100">
                    <div
                      className="absolute inset-y-0 left-0 rounded-full"
                      style={{
                        width: `${(server.cost / totalCost) * 100}%`,
                        backgroundColor: COLORS[i % COLORS.length],
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Cost by Team</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[250px] flex items-center justify-center">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={costByTeam}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {costByTeam.map((entry, index) => (
                      <Cell key={entry.name} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#fff',
                      border: '1px solid #e5e7eb',
                      borderRadius: '8px',
                    }}
                    formatter={(value: number) => [`$${value.toFixed(2)}`, 'Cost']}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex flex-wrap justify-center gap-4 mt-4">
              {costByTeam.map((team, i) => (
                <div key={team.name} className="flex items-center gap-2 text-sm">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: COLORS[i % COLORS.length] }}
                  />
                  <span className="text-gray-600">{team.name}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
