'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Save, Loader2, RefreshCw, Plus, Trash2, TestTube, Check, X } from 'lucide-react';
import {
  useSettings,
  useUpdateSettings,
  useSSOProviders,
  useCreateSSOProvider,
  useDeleteSSOProvider,
  useTelemetryConfigs,
  useCreateTelemetryConfig,
  useDeleteTelemetryConfig,
  useTestTelemetryConfig,
} from '@/lib/hooks/use-api';
import { useSWRConfig } from 'swr';
import { SSOProviderType, TelemetryExporterType, TelemetryProtocol } from '@/lib/api/client';

export default function SettingsPage() {
  const { mutate } = useSWRConfig();

  // Fetch data
  const { data: settings, isLoading: settingsLoading, error: settingsError } = useSettings();
  const { data: ssoData, isLoading: ssoLoading } = useSSOProviders(true);
  const { data: telemetryData, isLoading: telemetryLoading } = useTelemetryConfigs();

  // Mutations
  const { trigger: updateSettings, isMutating: isUpdating } = useUpdateSettings();
  const { trigger: createSSOProvider, isMutating: isCreatingSSO } = useCreateSSOProvider();
  const { trigger: deleteSSOProvider } = useDeleteSSOProvider();
  const { trigger: createTelemetryConfig, isMutating: isCreatingTelemetry } = useCreateTelemetryConfig();
  const { trigger: deleteTelemetryConfig } = useDeleteTelemetryConfig();
  const { trigger: testTelemetryConfig } = useTestTelemetryConfig();

  // Form state
  const [orgName, setOrgName] = useState('');
  const [billingEmail, setBillingEmail] = useState('');
  const [productionRPM, setProductionRPM] = useState(1000);
  const [sandboxRPM, setSandboxRPM] = useState(100);

  // SSO form state
  const [showSSOForm, setShowSSOForm] = useState(false);
  const [ssoType, setSSOType] = useState<SSOProviderType>('okta');
  const [ssoName, setSSOName] = useState('');
  const [ssoIssuerUrl, setSSOIssuerUrl] = useState('');
  const [ssoClientId, setSSOClientId] = useState('');
  const [ssoClientSecret, setSSOClientSecret] = useState('');

  // Telemetry form state
  const [showTelemetryForm, setShowTelemetryForm] = useState(false);
  const [telemetryName, setTelemetryName] = useState('');
  const [telemetryEndpoint, setTelemetryEndpoint] = useState('');
  const [telemetryProtocol, setTelemetryProtocol] = useState<TelemetryProtocol>('grpc');
  const [telemetryExporterType, setTelemetryExporterType] = useState<TelemetryExporterType>('otlp');

  // Test results
  const [testingConfigId, setTestingConfigId] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);

  // Saved state
  const [saved, setSaved] = useState(false);

  // Initialize form with settings data
  useEffect(() => {
    if (settings) {
      setOrgName(settings.org_name);
      setBillingEmail(settings.billing_email);
      setProductionRPM(settings.rate_limits.production_rpm);
      setSandboxRPM(settings.rate_limits.sandbox_rpm);
    }
  }, [settings]);

  const handleSaveSettings = async () => {
    try {
      await updateSettings({
        org_name: orgName,
        billing_email: billingEmail,
        rate_limits: {
          production_rpm: productionRPM,
          sandbox_rpm: sandboxRPM,
        },
      });
      mutate('settings');
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('Failed to save settings:', err);
    }
  };

  const handleCreateSSOProvider = async () => {
    if (!ssoName || !ssoIssuerUrl || !ssoClientId || !ssoClientSecret) return;
    try {
      await createSSOProvider({
        type: ssoType,
        name: ssoName,
        issuer_url: ssoIssuerUrl,
        client_id: ssoClientId,
        client_secret: ssoClientSecret,
      });
      mutate('sso-providers-all');
      setShowSSOForm(false);
      setSSOName('');
      setSSOIssuerUrl('');
      setSSOClientId('');
      setSSOClientSecret('');
    } catch (err) {
      console.error('Failed to create SSO provider:', err);
    }
  };

  const handleDeleteSSOProvider = async (providerId: string) => {
    try {
      await deleteSSOProvider(providerId);
      mutate('sso-providers-all');
    } catch (err) {
      console.error('Failed to delete SSO provider:', err);
    }
  };

  const handleCreateTelemetryConfig = async () => {
    if (!telemetryName || !telemetryEndpoint) return;
    try {
      await createTelemetryConfig({
        name: telemetryName,
        endpoint: telemetryEndpoint,
        protocol: telemetryProtocol,
        exporter_type: telemetryExporterType,
        enabled: true,
      });
      mutate('telemetry-configs');
      setShowTelemetryForm(false);
      setTelemetryName('');
      setTelemetryEndpoint('');
    } catch (err) {
      console.error('Failed to create telemetry config:', err);
    }
  };

  const handleDeleteTelemetryConfig = async (configId: string) => {
    try {
      await deleteTelemetryConfig(configId);
      mutate('telemetry-configs');
    } catch (err) {
      console.error('Failed to delete telemetry config:', err);
    }
  };

  const handleTestTelemetryConfig = async (configId: string) => {
    setTestingConfigId(configId);
    setTestResult(null);
    try {
      const result = await testTelemetryConfig(configId);
      setTestResult({ success: result.success, message: result.message });
    } catch (err) {
      setTestResult({ success: false, message: 'Connection test failed' });
    } finally {
      setTimeout(() => {
        setTestingConfigId(null);
        setTestResult(null);
      }, 3000);
    }
  };

  const isLoading = settingsLoading || ssoLoading || telemetryLoading;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-600" />
      </div>
    );
  }

  if (settingsError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-red-600">Failed to load settings</p>
        <Button variant="outline" onClick={() => window.location.reload()}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  const ssoProviders = ssoData?.providers || [];
  const telemetryConfigs = telemetryData?.configs || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500">Manage your GatewayOps configuration</p>
      </div>

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Organization</CardTitle>
            <CardDescription>Basic organization settings</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Organization Name
                </label>
                <input
                  type="text"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Billing Email
                </label>
                <input
                  type="email"
                  value={billingEmail}
                  onChange={(e) => setBillingEmail(e.target.value)}
                  className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Default Rate Limits</CardTitle>
            <CardDescription>Set default rate limits for new API keys</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Production (RPM)
                </label>
                <input
                  type="number"
                  value={productionRPM}
                  onChange={(e) => setProductionRPM(Number(e.target.value))}
                  className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Sandbox (RPM)
                </label>
                <input
                  type="number"
                  value={sandboxRPM}
                  onChange={(e) => setSandboxRPM(Number(e.target.value))}
                  className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>SSO Configuration</CardTitle>
              <CardDescription>Configure single sign-on for your organization</CardDescription>
            </div>
            {!showSSOForm && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowSSOForm(true)}
              >
                <Plus className="h-4 w-4 mr-2" />
                Add Provider
              </Button>
            )}
          </CardHeader>
          <CardContent className="space-y-4">
            {showSSOForm && (
              <div className="border rounded-lg p-4 space-y-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <h4 className="font-medium">New SSO Provider</h4>
                  <Button variant="ghost" size="icon" onClick={() => setShowSSOForm(false)}>
                    <X className="h-4 w-4" />
                  </Button>
                </div>
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Provider Type
                    </label>
                    <select
                      value={ssoType}
                      onChange={(e) => setSSOType(e.target.value as SSOProviderType)}
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    >
                      <option value="okta">Okta</option>
                      <option value="azure_ad">Azure AD</option>
                      <option value="google">Google Workspace</option>
                      <option value="onelogin">OneLogin</option>
                      <option value="auth0">Auth0</option>
                      <option value="oidc">Generic OIDC</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Display Name
                    </label>
                    <input
                      type="text"
                      value={ssoName}
                      onChange={(e) => setSSOName(e.target.value)}
                      placeholder="e.g., Company Okta"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Issuer URL
                    </label>
                    <input
                      type="url"
                      value={ssoIssuerUrl}
                      onChange={(e) => setSSOIssuerUrl(e.target.value)}
                      placeholder="https://your-domain.okta.com"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Client ID
                    </label>
                    <input
                      type="text"
                      value={ssoClientId}
                      onChange={(e) => setSSOClientId(e.target.value)}
                      placeholder="Enter client ID"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Client Secret
                    </label>
                    <input
                      type="password"
                      value={ssoClientSecret}
                      onChange={(e) => setSSOClientSecret(e.target.value)}
                      placeholder="Enter client secret"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                </div>
                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={() => setShowSSOForm(false)}>
                    Cancel
                  </Button>
                  <Button
                    onClick={handleCreateSSOProvider}
                    disabled={!ssoName || !ssoIssuerUrl || !ssoClientId || !ssoClientSecret || isCreatingSSO}
                    className="bg-indigo-600 hover:bg-indigo-700 text-white"
                  >
                    {isCreatingSSO ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      'Add Provider'
                    )}
                  </Button>
                </div>
              </div>
            )}

            {ssoProviders.length > 0 ? (
              <div className="space-y-2">
                {ssoProviders.map((provider) => (
                  <div
                    key={provider.id}
                    className="flex items-center justify-between border rounded-lg p-3"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{provider.name}</span>
                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                          provider.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                        }`}>
                          {provider.enabled ? 'Active' : 'Disabled'}
                        </span>
                      </div>
                      <p className="text-sm text-gray-500">
                        {provider.type.replace('_', ' ').toUpperCase()} - {provider.issuer_url}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleDeleteSSOProvider(provider.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            ) : !showSSOForm && (
              <p className="text-sm text-gray-500 text-center py-4">
                No SSO providers configured. Click &quot;Add Provider&quot; to set up SSO.
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>OpenTelemetry Export</CardTitle>
              <CardDescription>Export traces to external observability platforms</CardDescription>
            </div>
            {!showTelemetryForm && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowTelemetryForm(true)}
              >
                <Plus className="h-4 w-4 mr-2" />
                Add Exporter
              </Button>
            )}
          </CardHeader>
          <CardContent className="space-y-4">
            {showTelemetryForm && (
              <div className="border rounded-lg p-4 space-y-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <h4 className="font-medium">New Telemetry Exporter</h4>
                  <Button variant="ghost" size="icon" onClick={() => setShowTelemetryForm(false)}>
                    <X className="h-4 w-4" />
                  </Button>
                </div>
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Name
                    </label>
                    <input
                      type="text"
                      value={telemetryName}
                      onChange={(e) => setTelemetryName(e.target.value)}
                      placeholder="e.g., Production Jaeger"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Exporter Type
                    </label>
                    <select
                      value={telemetryExporterType}
                      onChange={(e) => setTelemetryExporterType(e.target.value as TelemetryExporterType)}
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    >
                      <option value="otlp">OTLP</option>
                      <option value="jaeger">Jaeger</option>
                      <option value="zipkin">Zipkin</option>
                      <option value="datadog">Datadog</option>
                      <option value="newrelic">New Relic</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Endpoint
                    </label>
                    <input
                      type="url"
                      value={telemetryEndpoint}
                      onChange={(e) => setTelemetryEndpoint(e.target.value)}
                      placeholder="https://otel-collector:4317"
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Protocol
                    </label>
                    <select
                      value={telemetryProtocol}
                      onChange={(e) => setTelemetryProtocol(e.target.value as TelemetryProtocol)}
                      className="w-full h-10 rounded-md border border-gray-200 px-3 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    >
                      <option value="grpc">gRPC</option>
                      <option value="http">HTTP/protobuf</option>
                    </select>
                  </div>
                </div>
                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={() => setShowTelemetryForm(false)}>
                    Cancel
                  </Button>
                  <Button
                    onClick={handleCreateTelemetryConfig}
                    disabled={!telemetryName || !telemetryEndpoint || isCreatingTelemetry}
                    className="bg-indigo-600 hover:bg-indigo-700 text-white"
                  >
                    {isCreatingTelemetry ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      'Add Exporter'
                    )}
                  </Button>
                </div>
              </div>
            )}

            {telemetryConfigs.length > 0 ? (
              <div className="space-y-2">
                {telemetryConfigs.map((config) => (
                  <div
                    key={config.id}
                    className="flex items-center justify-between border rounded-lg p-3"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{config.name}</span>
                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                          config.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                        }`}>
                          {config.enabled ? 'Active' : 'Disabled'}
                        </span>
                        {testingConfigId === config.id && testResult && (
                          <span className={`text-xs px-2 py-0.5 rounded-full ${
                            testResult.success ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                          }`}>
                            {testResult.success ? 'Connected' : 'Failed'}
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-500">
                        {config.exporter_type.toUpperCase()} via {config.protocol.toUpperCase()} - {config.endpoint}
                      </p>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleTestTelemetryConfig(config.id)}
                        disabled={testingConfigId === config.id}
                      >
                        {testingConfigId === config.id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <TestTube className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-red-600 hover:text-red-700"
                        onClick={() => handleDeleteTelemetryConfig(config.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            ) : !showTelemetryForm && (
              <p className="text-sm text-gray-500 text-center py-4">
                No telemetry exporters configured. Click &quot;Add Exporter&quot; to set up OTLP export.
              </p>
            )}
          </CardContent>
        </Card>

        <div className="flex justify-end items-center gap-4">
          {saved && (
            <span className="flex items-center text-green-600 text-sm">
              <Check className="h-4 w-4 mr-1" />
              Settings saved
            </span>
          )}
          <Button
            onClick={handleSaveSettings}
            disabled={isUpdating}
            className="bg-indigo-600 hover:bg-indigo-700 text-white"
          >
            {isUpdating ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            Save Changes
          </Button>
        </div>
      </div>
    </div>
  );
}
