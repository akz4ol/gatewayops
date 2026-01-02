'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { AuthContextType, AuthState, User, Session, LoginCredentials, SSOProvider } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://gatewayops-api.fly.dev';

const initialState: AuthState = {
  user: null,
  session: null,
  isLoading: true,
  isAuthenticated: false,
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const PUBLIC_PATHS = ['/login', '/auth/callback'];

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>(initialState);
  const router = useRouter();
  const pathname = usePathname();

  // Check for existing session on mount
  useEffect(() => {
    checkSession();
  }, []);

  // Redirect logic
  useEffect(() => {
    if (state.isLoading) return;

    const isPublicPath = PUBLIC_PATHS.some(path => pathname?.startsWith(path));

    if (!state.isAuthenticated && !isPublicPath) {
      router.push('/login');
    } else if (state.isAuthenticated && pathname === '/login') {
      router.push('/');
    }
  }, [state.isAuthenticated, state.isLoading, pathname, router]);

  const checkSession = useCallback(async () => {
    try {
      const token = localStorage.getItem('session_token');
      if (!token) {
        setState(prev => ({ ...prev, isLoading: false }));
        return;
      }

      // Validate session with backend
      const response = await fetch(`${API_BASE}/v1/rbac/me`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        // For demo mode, create user from permissions response
        const user: User = {
          id: data.user_id || '00000000-0000-0000-0000-000000000001',
          email: 'admin@demo.gatewayops.io',
          name: 'Demo Admin',
          org_id: '00000000-0000-0000-0000-000000000001',
          org_name: 'GatewayOps Demo',
        };

        setState({
          user,
          session: { id: '', user_id: user.id, access_token: token, expires_at: '' },
          isLoading: false,
          isAuthenticated: true,
        });
      } else {
        // Invalid token, clear it
        localStorage.removeItem('session_token');
        setState(prev => ({ ...prev, isLoading: false }));
      }
    } catch (error) {
      console.error('Session check failed:', error);
      setState(prev => ({ ...prev, isLoading: false }));
    }
  }, []);

  const login = async (credentials: LoginCredentials) => {
    setState(prev => ({ ...prev, isLoading: true }));

    try {
      // For demo mode, accept any credentials and create a demo session
      // In production, this would call a real auth endpoint
      if (credentials.email && credentials.password) {
        // Simulate API call delay
        await new Promise(resolve => setTimeout(resolve, 500));

        // Demo user
        const user: User = {
          id: '00000000-0000-0000-0000-000000000001',
          email: credentials.email,
          name: credentials.email.split('@')[0],
          org_id: '00000000-0000-0000-0000-000000000001',
          org_name: 'GatewayOps Demo',
        };

        // Generate demo token
        const token = 'demo_' + btoa(credentials.email + ':' + Date.now());
        localStorage.setItem('session_token', token);

        setState({
          user,
          session: { id: '', user_id: user.id, access_token: token, expires_at: '' },
          isLoading: false,
          isAuthenticated: true,
        });

        router.push('/');
      }
    } catch (error) {
      setState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  const loginWithSSO = (providerId: string) => {
    // Redirect to SSO authorization endpoint
    const redirectUrl = encodeURIComponent(window.location.origin + '/auth/callback');
    window.location.href = `${API_BASE}/v1/sso/authorize/${providerId}?redirect_url=${redirectUrl}`;
  };

  const logout = async () => {
    try {
      const token = localStorage.getItem('session_token');
      if (token) {
        // Call logout endpoint
        await fetch(`${API_BASE}/v1/sso/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }).catch(() => {}); // Ignore errors
      }
    } finally {
      localStorage.removeItem('session_token');
      setState({
        user: null,
        session: null,
        isLoading: false,
        isAuthenticated: false,
      });
      router.push('/login');
    }
  };

  const refreshSession = async () => {
    await checkSession();
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        loginWithSSO,
        logout,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Hook to get SSO providers
export function useSSOProviders() {
  const [providers, setProviders] = useState<SSOProvider[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProviders = async () => {
      try {
        const response = await fetch(`${API_BASE}/v1/sso/providers`);
        if (response.ok) {
          const data = await response.json();
          setProviders(data.providers?.filter((p: SSOProvider) => p.enabled) || []);
        }
      } catch (error) {
        console.error('Failed to fetch SSO providers:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchProviders();
  }, []);

  return { providers, loading };
}
