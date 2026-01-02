'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Bell, AlertTriangle, CheckCircle, Clock } from 'lucide-react';

const alertRules = [
  {
    id: 1,
    name: 'High Error Rate',
    metric: 'error_rate',
    condition: '> 5%',
    window: '5 min',
    severity: 'critical',
    enabled: true,
  },
  {
    id: 2,
    name: 'Latency Spike',
    metric: 'p99_latency',
    condition: '> 500ms',
    window: '5 min',
    severity: 'warning',
    enabled: true,
  },
  {
    id: 3,
    name: 'Cost Anomaly',
    metric: 'hourly_cost',
    condition: '> $50',
    window: '1 hour',
    severity: 'warning',
    enabled: false,
  },
];

const activeAlerts = [
  {
    id: 1,
    rule: 'High Error Rate',
    message: 'Error rate at 7.2% for github server',
    severity: 'critical',
    status: 'firing',
    startedAt: '10 min ago',
  },
  {
    id: 2,
    rule: 'Latency Spike',
    message: 'P99 latency at 892ms for database server',
    severity: 'warning',
    status: 'firing',
    startedAt: '25 min ago',
  },
];

const alertHistory = [
  {
    id: 1,
    rule: 'High Error Rate',
    status: 'resolved',
    duration: '15 min',
    resolvedAt: '2 hours ago',
  },
  {
    id: 2,
    rule: 'Cost Anomaly',
    status: 'resolved',
    duration: '45 min',
    resolvedAt: '5 hours ago',
  },
  {
    id: 3,
    rule: 'Latency Spike',
    status: 'resolved',
    duration: '8 min',
    resolvedAt: '1 day ago',
  },
];

export default function AlertsPage() {
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
                <p className="text-2xl font-bold text-gray-900">2</p>
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
                <p className="text-2xl font-bold text-gray-900">15</p>
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
                <p className="text-2xl font-bold text-gray-900">3</p>
                <p className="text-sm text-gray-500">Alert Rules</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {activeAlerts.length > 0 && (
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-lg font-medium text-red-900">Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {activeAlerts.map((alert) => (
                <div key={alert.id} className="flex items-center justify-between rounded-lg bg-white border border-red-200 p-4">
                  <div className="flex items-center gap-4">
                    <div className={`rounded-full p-2 ${
                      alert.severity === 'critical' ? 'bg-red-100' : 'bg-yellow-100'
                    }`}>
                      <AlertTriangle className={`h-4 w-4 ${
                        alert.severity === 'critical' ? 'text-red-600' : 'text-yellow-600'
                      }`} />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{alert.rule}</p>
                      <p className="text-sm text-gray-600">{alert.message}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="flex items-center gap-1 text-sm text-gray-500">
                      <Clock className="h-4 w-4" />
                      {alert.startedAt}
                    </span>
                    <span className="rounded-full bg-red-100 px-2 py-1 text-xs font-medium text-red-700">
                      {alert.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Alert Rules</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {alertRules.map((rule) => (
                <div key={rule.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-4">
                    <div className={`h-3 w-3 rounded-full ${rule.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
                    <div>
                      <p className="font-medium text-gray-900">{rule.name}</p>
                      <p className="text-sm text-gray-500">
                        {rule.metric} {rule.condition} over {rule.window}
                      </p>
                    </div>
                  </div>
                  <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                    rule.severity === 'critical' ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'
                  }`}>
                    {rule.severity}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Recent History</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {alertHistory.map((alert) => (
                <div key={alert.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-4">
                    <div className="rounded-full bg-green-100 p-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{alert.rule}</p>
                      <p className="text-sm text-gray-500">Duration: {alert.duration}</p>
                    </div>
                  </div>
                  <span className="text-sm text-gray-500">{alert.resolvedAt}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
