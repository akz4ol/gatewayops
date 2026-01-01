'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Shield, AlertTriangle, CheckCircle, XCircle } from 'lucide-react';

const policies = [
  {
    id: 1,
    name: 'Default Injection Detection',
    mode: 'block',
    sensitivity: 'high',
    patterns: 12,
    blocked: 47,
    enabled: true,
  },
  {
    id: 2,
    name: 'Code Execution Prevention',
    mode: 'warn',
    sensitivity: 'moderate',
    patterns: 8,
    blocked: 23,
    enabled: true,
  },
  {
    id: 3,
    name: 'Data Exfiltration Monitor',
    mode: 'log',
    sensitivity: 'low',
    patterns: 5,
    blocked: 0,
    enabled: false,
  },
];

const recentDetections = [
  {
    id: 1,
    type: 'Prompt Injection',
    severity: 'high',
    server: 'filesystem',
    action: 'blocked',
    time: '5 min ago',
  },
  {
    id: 2,
    type: 'Jailbreak Attempt',
    severity: 'critical',
    server: 'database',
    action: 'blocked',
    time: '12 min ago',
  },
  {
    id: 3,
    type: 'Sensitive Data Request',
    severity: 'medium',
    server: 'github',
    action: 'warned',
    time: '1 hour ago',
  },
];

export default function SafetyPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Safety & Security</h1>
        <p className="text-gray-500">Monitor and configure prompt injection detection</p>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-green-100 p-3">
                <Shield className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">98.2%</p>
                <p className="text-sm text-gray-500">Detection Rate</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-red-100 p-3">
                <XCircle className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">47</p>
                <p className="text-sm text-gray-500">Blocked Today</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-yellow-100 p-3">
                <AlertTriangle className="h-5 w-5 text-yellow-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">12</p>
                <p className="text-sm text-gray-500">Warnings</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-blue-100 p-3">
                <CheckCircle className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">3</p>
                <p className="text-sm text-gray-500">Active Policies</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Safety Policies</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {policies.map((policy) => (
                <div key={policy.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-4">
                    <div className={`h-3 w-3 rounded-full ${policy.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
                    <div>
                      <p className="font-medium text-gray-900">{policy.name}</p>
                      <p className="text-sm text-gray-500">
                        {policy.patterns} patterns | Mode: {policy.mode}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-medium text-gray-900">{policy.blocked}</p>
                    <p className="text-sm text-gray-500">blocked</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Recent Detections</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentDetections.map((detection) => (
                <div key={detection.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-4">
                    <div className={`rounded-full p-2 ${
                      detection.severity === 'critical' ? 'bg-red-100' :
                      detection.severity === 'high' ? 'bg-orange-100' : 'bg-yellow-100'
                    }`}>
                      <AlertTriangle className={`h-4 w-4 ${
                        detection.severity === 'critical' ? 'text-red-600' :
                        detection.severity === 'high' ? 'text-orange-600' : 'text-yellow-600'
                      }`} />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{detection.type}</p>
                      <p className="text-sm text-gray-500">{detection.server} | {detection.time}</p>
                    </div>
                  </div>
                  <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                    detection.action === 'blocked' ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'
                  }`}>
                    {detection.action}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
