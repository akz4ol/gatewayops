export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  org_id: string;
  org_name?: string;
}

export interface Session {
  id: string;
  user_id: string;
  access_token: string;
  expires_at: string;
}

export interface SSOProvider {
  id: string;
  type: string;
  name: string;
  enabled: boolean;
  issuer_url: string;
}

export interface AuthState {
  user: User | null;
  session: Session | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  loginWithSSO: (providerId: string) => void;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}
