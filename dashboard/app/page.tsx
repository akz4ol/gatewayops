'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatsCard } from '@/components/dashboard/stats-card';
import { RequestsChart } from '@/components/dashboard/requests-chart';
import { RecentTraces } from '@/components/dashboard/recent-traces';
import { TopServers } from '@/components/dashboard/top-servers';
import { Activity, DollarSign, Clock, AlertTriangle } from 'lucide-react';

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500">Overview of your MCP Gateway activity</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard
          title="Total Requests"
          value="1.2M"
          change={12.5}
          icon={Activity}
          description="Last 30 days"
        />
        <StatsCard
          title="Total Cost"
          value="$4,231.89"
          change={-3.2}
          icon={DollarSign}
          description="Last 30 days"
        />
        <StatsCard
          title="Avg Latency"
          value="127ms"
          change={-8.1}
          icon={Clock}
          description="P95 latency"
        />
        <StatsCard
          title="Error Rate"
          value="0.12%"
          change={0.02}
          icon={AlertTriangle}
          description="Last 24 hours"
          changeType="negative"
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-7">
        <Card className="lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-lg font-medium">Request Volume</CardTitle>
          </CardHeader>
          <CardContent>
            <RequestsChart />
          </CardContent>
        </Card>

        <Card className="lg:col-span-3">
          <CardHeader>
            <CardTitle className="text-lg font-medium">Top MCP Servers</CardTitle>
          </CardHeader>
          <CardContent>
            <TopServers />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Recent Traces</CardTitle>
        </CardHeader>
        <CardContent>
          <RecentTraces />
        </CardContent>
      </Card>
    </div>
  );
}
