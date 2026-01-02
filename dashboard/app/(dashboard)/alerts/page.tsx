'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Bell, AlertTriangle, CheckCircle, Clock, Loader2 } from 'lucide-react';
import { useAlertRules, useActiveAlerts, useAlerts } from '@/lib/hooks/use-api';
import { formatDistanceToNow } from 'date-fns';
import { AlertRule, Alert, AlertSeverity, AlertCondition, AlertMetric } from '@/lib/api/client';

// Helper to format condition for display
function formatCondition(condition: AlertCondition, threshold: number, metric: AlertMetric): string {
  const conditionSymbols: Record<AlertCondition, string> = {
    gt: '>',
    lt: '<',
    gte: '>=',
    lte: '<=',
    eq: '=',
    neq: '!=',
  };

  const symbol = conditionSymbols[condition] || condition;

  // Format threshold based on metric type
  if (metric.includes('rate') || metric === 'error_rate') {
    return `${symbol} ${threshold}%`;
  } else if (metric.includes('latency')) {
    return `${symbol} ${threshold}ms`;
  } else if (metric.includes('cost')) {
    return `${symbol} $${threshold}`;
  }
  return `${symbol} ${threshold}`;
}

// Helper to format window
function formatWindow(minutes: number): string {
  if (minutes >= 60) {
    const hours = Math.floor(minutes / 60);
    return `${hours} hour${hours > 1 ? 's' : ''}`;
  }
  return `${minutes} min`;
}

// Helper to format time ago
function formatTimeAgo(dateString: string): string {
  try {
    return formatDistanceToNow(new Date(dateString), { addSuffix: true });
  } catch {
    return dateString;
  }
}

// Loading skeleton component
function LoadingSkeleton() {
  return (
    <div className="flex items-center justify-center py-8">
      <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
    </div>
  );
}

// Empty state component
function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-8 text-gray-500">
      <Bell className="h-8 w-8 mb-2 opacity-50" />
      <p>{message}</p>
    </div>
  );
}

export default function AlertsPage() {
  const { data: rulesData, isLoading: rulesLoading, error: rulesError } = useAlertRules();
  const { data: activeAlertsData, isLoading: activeLoading, error: activeError } = useActiveAlerts();
  const { data: alertsHistory, isLoading: historyLoading } = useAlerts({
    statuses: ['resolved'],
    limit: 10
  });

  const alertRules = rulesData?.rules || [];
  const activeAlerts = activeAlertsData?.alerts || [];
  const resolvedAlerts = alertsHistory?.alerts || [];

  // Calculate stats
  const activeCount = activeAlerts.length;
  const rulesCount = alertRules.length;
  const resolvedTodayCount = resolvedAlerts.filter(a => {
    if (!a.resolved_at) return false;
    const resolvedDate = new Date(a.resolved_at);
    const today = new Date();
    return resolvedDate.toDateString() === today.toDateString();
  }).length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Alerts</h1>
        <p className="text-gray-500">Configure alert rules and view active alerts</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-red-100 p-3">
                <AlertTriangle className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">
                  {activeLoading ? '-' : activeCount}
                </p>
                <p className="text-sm text-gray-500">Active Alerts</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-green-100 p-3">
                <CheckCircle className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">
                  {historyLoading ? '-' : resolvedTodayCount}
                </p>
                <p className="text-sm text-gray-500">Resolved Today</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-blue-100 p-3">
                <Bell className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">
                  {rulesLoading ? '-' : rulesCount}
                </p>
                <p className="text-sm text-gray-500">Alert Rules</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Active Alerts Section */}
      {activeLoading ? (
        <Card className="border-gray-200">
          <CardHeader>
            <CardTitle className="text-lg font-medium">Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <LoadingSkeleton />
          </CardContent>
        </Card>
      ) : activeError ? (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="pt-6">
            <p className="text-red-600">Failed to load active alerts</p>
          </CardContent>
        </Card>
      ) : activeAlerts.length > 0 ? (
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-lg font-medium text-red-900">Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {activeAlerts.map((alert) => (
                <AlertCard key={alert.id} alert={alert} />
              ))}
            </div>
          </CardContent>
        </Card>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Alert Rules */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Alert Rules</CardTitle>
          </CardHeader>
          <CardContent>
            {rulesLoading ? (
              <LoadingSkeleton />
            ) : rulesError ? (
              <p className="text-red-600 py-4">Failed to load alert rules</p>
            ) : alertRules.length === 0 ? (
              <EmptyState message="No alert rules configured" />
            ) : (
              <div className="space-y-4">
                {alertRules.map((rule) => (
                  <RuleCard key={rule.id} rule={rule} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent History */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Recent History</CardTitle>
          </CardHeader>
          <CardContent>
            {historyLoading ? (
              <LoadingSkeleton />
            ) : resolvedAlerts.length === 0 ? (
              <EmptyState message="No resolved alerts" />
            ) : (
              <div className="space-y-4">
                {resolvedAlerts.map((alert) => (
                  <HistoryCard key={alert.id} alert={alert} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

// Alert card component for active alerts
function AlertCard({ alert }: { alert: Alert }) {
  return (
    <div className="flex items-center justify-between rounded-lg bg-white border border-red-200 p-4">
      <div className="flex items-center gap-4">
        <div className={`rounded-full p-2 ${
          alert.severity === 'critical' ? 'bg-red-100' : 'bg-yellow-100'
        }`}>
          <AlertTriangle className={`h-4 w-4 ${
            alert.severity === 'critical' ? 'text-red-600' : 'text-yellow-600'
          }`} />
        </div>
        <div>
          <p className="font-medium text-gray-900">{alert.message}</p>
          <p className="text-sm text-gray-600">
            Value: {alert.value.toFixed(2)} (threshold: {alert.threshold})
          </p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <span className="flex items-center gap-1 text-sm text-gray-500">
          <Clock className="h-4 w-4" />
          {formatTimeAgo(alert.started_at)}
        </span>
        <SeverityBadge severity={alert.severity} />
      </div>
    </div>
  );
}

// Rule card component
function RuleCard({ rule }: { rule: AlertRule }) {
  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <div className={`h-3 w-3 rounded-full ${rule.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
        <div>
          <p className="font-medium text-gray-900">{rule.name}</p>
          <p className="text-sm text-gray-500">
            {rule.metric} {formatCondition(rule.condition, rule.threshold, rule.metric)} over {formatWindow(rule.window_minutes)}
          </p>
        </div>
      </div>
      <SeverityBadge severity={rule.severity} />
    </div>
  );
}

// History card component for resolved alerts
function HistoryCard({ alert }: { alert: Alert }) {
  const duration = alert.resolved_at && alert.started_at
    ? calculateDuration(alert.started_at, alert.resolved_at)
    : 'Unknown';

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <div className="rounded-full bg-green-100 p-2">
          <CheckCircle className="h-4 w-4 text-green-600" />
        </div>
        <div>
          <p className="font-medium text-gray-900">{alert.message}</p>
          <p className="text-sm text-gray-500">Duration: {duration}</p>
        </div>
      </div>
      <span className="text-sm text-gray-500">
        {alert.resolved_at ? formatTimeAgo(alert.resolved_at) : 'Unknown'}
      </span>
    </div>
  );
}

// Severity badge component
function SeverityBadge({ severity }: { severity: AlertSeverity }) {
  const colors: Record<AlertSeverity, string> = {
    critical: 'bg-red-100 text-red-700',
    warning: 'bg-yellow-100 text-yellow-700',
    info: 'bg-blue-100 text-blue-700',
  };

  return (
    <span className={`rounded-full px-2 py-1 text-xs font-medium ${colors[severity] || colors.info}`}>
      {severity}
    </span>
  );
}

// Helper to calculate duration between two dates
function calculateDuration(start: string, end: string): string {
  const startDate = new Date(start);
  const endDate = new Date(end);
  const diffMs = endDate.getTime() - startDate.getTime();

  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 60) {
    return `${minutes} min`;
  }

  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (hours < 24) {
    return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
  }

  const days = Math.floor(hours / 24);
  return `${days}d ${hours % 24}h`;
}
