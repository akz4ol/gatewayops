const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api';

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export class ApiClient {
  private baseUrl: string;
  private apiKey?: string;

  constructor(baseUrl: string = API_BASE, apiKey?: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  private async request<T>(
    method: string,
    path: string,
    options?: { body?: unknown; params?: Record<string, string> }
  ): Promise<T> {
    const url = new URL(path, this.baseUrl);
    if (options?.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        url.searchParams.set(key, value);
      });
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.apiKey) {
      headers['Authorization'] = `Bearer ${this.apiKey}`;
    }

    const response = await fetch(url.toString(), {
      method,
      headers,
      body: options?.body ? JSON.stringify(options.body) : undefined,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || 'Request failed');
    }

    return data;
  }

  get<T>(path: string, params?: Record<string, string>): Promise<T> {
    return this.request<T>('GET', path, { params });
  }

  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('POST', path, { body });
  }

  put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PUT', path, { body });
  }

  delete<T>(path: string): Promise<T> {
    return this.request<T>('DELETE', path);
  }
}

export const api = new ApiClient();
