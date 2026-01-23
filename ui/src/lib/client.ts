import { createPromiseClient, Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CatalogService } from '@gen/catalog/v1/catalog_connect';

// Session management
let sessionId: string | null = null;

export function getSessionId(): string | null {
  return sessionId;
}

export function clearSession(): void {
  sessionId = null;
}

// Interceptor to handle session IDs
const sessionInterceptor: Interceptor = (next) => async (req) => {
  // Add session ID to request if we have one
  if (sessionId) {
    req.header.set('X-Session-ID', sessionId);
  }

  const res = await next(req);

  // Store session ID from response
  const newSessionId = res.header.get('X-Session-ID');
  if (newSessionId) {
    sessionId = newSessionId;
  }

  return res;
};

// Get base URL from environment variable or default to localhost
const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

const transport = createConnectTransport({
  baseUrl,
  interceptors: [sessionInterceptor],
});

export const catalogClient = createPromiseClient(CatalogService, transport);
