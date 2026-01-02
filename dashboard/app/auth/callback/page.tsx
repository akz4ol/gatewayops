'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Loader2, AlertCircle } from 'lucide-react';

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      // Check for error from SSO provider
      const errorParam = searchParams.get('error');
      if (errorParam) {
        setError(searchParams.get('error_description') || 'Authentication failed');
        return;
      }

      // Get the access token from the URL or response
      // The backend SSO callback should have set a session cookie or returned a token
      const token = searchParams.get('token') || searchParams.get('session');

      if (token) {
        // Store the token
        localStorage.setItem('session_token', token);
        router.push('/');
      } else {
        // If no token in URL, the backend might have set a cookie
        // Try to verify the session
        try {
          const response = await fetch(
            `${process.env.NEXT_PUBLIC_API_URL || 'https://gatewayops-api.fly.dev'}/v1/rbac/me`,
            { credentials: 'include' }
          );

          if (response.ok) {
            router.push('/');
          } else {
            setError('Failed to complete authentication');
          }
        } catch (err) {
          setError('Failed to verify authentication');
        }
      }
    };

    handleCallback();
  }, [searchParams, router]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md w-full p-6 text-center">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <AlertCircle className="w-6 h-6 text-red-600" />
          </div>
          <h1 className="text-xl font-semibold text-gray-900 mb-2">Authentication Failed</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <button
            onClick={() => router.push('/login')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Back to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600 mx-auto mb-4" />
        <p className="text-gray-600">Completing sign in...</p>
      </div>
    </div>
  );
}
