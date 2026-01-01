'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatsCard } from '@/components/dashboard/stats-card';
import { RequestsChart } from '@/components/dashboard/requests-chart';
import { RecentTraces } from '@/components/dashboard/recent-traces';
import { TopServers } from '@/components/dashboard/top-servers';
import { Activity, DollarSign, Clock, AlertTriangle, Loader2 } from 'lucide-react';
import { useOverview } from '@/lib/hooks/use-api';

export default function DashboardPage() {
  const { data: overview, isLoading, error } = useOverview();

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
          <p className="text-gray-600">Failed to load dashboard data</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500">Overview of your MCP Gateway activity</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard
          title="Total Requests"
          value={overview?.total_requests.formatted || '0'}
          change={overview?.total_requests.change || 0}
          icon={Activity}
          description={`Last ${overview?.total_requests.period || '30d'}`}
        />
        <StatsCard
          title="Total Cost"
          value={overview?.total_cost.formatted || '$0'}
          change={overview?.total_cost.change || 0}
          icon={DollarSign}
          description={`Last ${overview?.total_cost.period || '30d'}`}
        />
        <StatsCard
          title="Avg Latency"
          value={overview?.avg_latency.formatted || '0ms'}
          change={overview?.avg_latency.change || 0}
          icon={Clock}
          description={overview?.avg_latency.percentile || 'P95 latency'}
        />
        <StatsCard
          title="Error Rate"
          value={overview?.error_rate.formatted || '0%'}
          change={overview?.error_rate.change || 0}
          icon={AlertTriangle}
          description={`Last ${overview?.error_rate.period || '24h'}`}
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
