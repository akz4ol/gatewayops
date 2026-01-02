'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Shield, AlertTriangle, CheckCircle, XCircle, Loader2 } from 'lucide-react';
import { useSafetyPolicies, useDetections, useSafetySummary } from '@/lib/hooks/use-api';
import { formatDistanceToNow } from 'date-fns';
import { SafetyPolicy, InjectionDetection, SafetyMode, DetectionSeverity } from '@/lib/api/client';

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
      <Shield className="h-8 w-8 mb-2 opacity-50" />
      <p>{message}</p>
    </div>
  );
}

export default function SafetyPage() {
  const { data: policiesData, isLoading: policiesLoading, error: policiesError } = useSafetyPolicies();
  const { data: detectionsData, isLoading: detectionsLoading, error: detectionsError } = useDetections({ limit: 10 });
  const { data: summaryData, isLoading: summaryLoading } = useSafetySummary();

  const policies = policiesData?.policies || [];
  const detections = detectionsData?.detections || [];
  const summary = summaryData;

  // Calculate stats from summary or defaults
  const blockedToday = summary?.by_action?.block || 0;
  const warnedToday = summary?.by_action?.warn || 0;
  const activePolicies = policies.filter(p => p.enabled).length;

  // Calculate detection rate (mock for now - would need historical data)
  const detectionRate = summary?.total_detections ? 98.2 : 0;

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
                <p className="text-2xl font-bold text-gray-900">
                  {summaryLoading ? '-' : `${detectionRate}%`}
                </p>
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
                <p className="text-2xl font-bold text-gray-900">
                  {summaryLoading ? '-' : blockedToday}
                </p>
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
                <p className="text-2xl font-bold text-gray-900">
                  {summaryLoading ? '-' : warnedToday}
                </p>
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
                <p className="text-2xl font-bold text-gray-900">
                  {policiesLoading ? '-' : activePolicies}
                </p>
                <p className="text-sm text-gray-500">Active Policies</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Safety Policies */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Safety Policies</CardTitle>
          </CardHeader>
          <CardContent>
            {policiesLoading ? (
              <LoadingSkeleton />
            ) : policiesError ? (
              <p className="text-red-600 py-4">Failed to load safety policies</p>
            ) : policies.length === 0 ? (
              <EmptyState message="No safety policies configured" />
            ) : (
              <div className="space-y-4">
                {policies.map((policy) => (
                  <PolicyCard key={policy.id} policy={policy} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent Detections */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Recent Detections</CardTitle>
          </CardHeader>
          <CardContent>
            {detectionsLoading ? (
              <LoadingSkeleton />
            ) : detectionsError ? (
              <p className="text-red-600 py-4">Failed to load detections</p>
            ) : detections.length === 0 ? (
              <EmptyState message="No detections recorded" />
            ) : (
              <div className="space-y-4">
                {detections.map((detection) => (
                  <DetectionCard key={detection.id} detection={detection} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

// Policy card component
function PolicyCard({ policy }: { policy: SafetyPolicy }) {
  const patternCount = (policy.patterns?.block?.length || 0) + (policy.patterns?.allow?.length || 0);

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <div className={`h-3 w-3 rounded-full ${policy.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
        <div>
          <p className="font-medium text-gray-900">{policy.name}</p>
          <p className="text-sm text-gray-500">
            {patternCount} patterns | Mode: {policy.mode}
          </p>
        </div>
      </div>
      <div className="text-right">
        <ModeBadge mode={policy.mode} />
        <p className="text-xs text-gray-500 mt-1">{policy.sensitivity}</p>
      </div>
    </div>
  );
}

// Detection card component
function DetectionCard({ detection }: { detection: InjectionDetection }) {
  const typeLabels: Record<string, string> = {
    prompt_injection: 'Prompt Injection',
    pii: 'PII Detected',
    secret: 'Secret Detected',
    malicious: 'Malicious Content',
  };

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <SeverityIcon severity={detection.severity} />
        <div>
          <p className="font-medium text-gray-900">
            {typeLabels[detection.type] || detection.type}
          </p>
          <p className="text-sm text-gray-500">
            {detection.mcp_server || 'Unknown server'} | {formatTimeAgo(detection.created_at)}
          </p>
        </div>
      </div>
      <ActionBadge action={detection.action_taken} />
    </div>
  );
}

// Severity icon component
function SeverityIcon({ severity }: { severity: DetectionSeverity }) {
  const colors: Record<DetectionSeverity, { bg: string; text: string }> = {
    critical: { bg: 'bg-red-100', text: 'text-red-600' },
    high: { bg: 'bg-orange-100', text: 'text-orange-600' },
    medium: { bg: 'bg-yellow-100', text: 'text-yellow-600' },
    low: { bg: 'bg-blue-100', text: 'text-blue-600' },
  };

  const color = colors[severity] || colors.medium;

  return (
    <div className={`rounded-full p-2 ${color.bg}`}>
      <AlertTriangle className={`h-4 w-4 ${color.text}`} />
    </div>
  );
}

// Mode badge component
function ModeBadge({ mode }: { mode: SafetyMode }) {
  const colors: Record<SafetyMode, string> = {
    block: 'bg-red-100 text-red-700',
    warn: 'bg-yellow-100 text-yellow-700',
    log: 'bg-gray-100 text-gray-700',
  };

  return (
    <span className={`rounded-full px-2 py-1 text-xs font-medium ${colors[mode] || colors.log}`}>
      {mode}
    </span>
  );
}

// Action badge component
function ActionBadge({ action }: { action: SafetyMode }) {
  const labels: Record<SafetyMode, string> = {
    block: 'blocked',
    warn: 'warned',
    log: 'logged',
  };

  const colors: Record<SafetyMode, string> = {
    block: 'bg-red-100 text-red-700',
    warn: 'bg-yellow-100 text-yellow-700',
    log: 'bg-gray-100 text-gray-700',
  };

  return (
    <span className={`rounded-full px-2 py-1 text-xs font-medium ${colors[action] || colors.log}`}>
      {labels[action] || action}
    </span>
  );
}
