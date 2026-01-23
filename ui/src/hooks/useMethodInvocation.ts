import { useState } from 'react';
import { catalogClient } from '@/lib/client';
import { Transport } from '@gen/catalog/v1/catalog_pb';
import type { MethodInfo } from '@/lib/types';

interface InvocationOptions {
  selectedService: string | null;
  currentMethod: MethodInfo | null;
  targetEndpoint: string;
  useTls: boolean;
  transport: Transport;
}

interface UseMethodInvocationReturn {
  response: Record<string, unknown> | undefined;
  responseTime: number | undefined;
  error: string | undefined;
  loading: boolean;
  invokeMethod: (request: Record<string, unknown>) => Promise<void>;
  clearResponse: () => void;
}

/**
 * Hook for managing method invocation and response handling
 */
export function useMethodInvocation(options: InvocationOptions): UseMethodInvocationReturn {
  const [response, setResponse] = useState<Record<string, unknown> | undefined>();
  const [responseTime, setResponseTime] = useState<number | undefined>();
  const [error, setError] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);

  const invokeMethod = async (request: Record<string, unknown>) => {
    const { selectedService, currentMethod, targetEndpoint, useTls, transport } = options;

    if (!selectedService || !currentMethod) {
      setError('No method selected');
      return;
    }

    setLoading(true);
    setError(undefined);
    setResponse(undefined);
    setResponseTime(undefined);

    const startTime = performance.now();

    try {
      // For Connect protocol (HTTP), make direct browser request
      // For gRPC, use backend proxy (browsers can't do HTTP/2 + binary protobuf)
      if (transport === Transport.CONNECT) {
        const protocol = useTls ? 'https' : 'http';
        const url = `${protocol}://${targetEndpoint}/${selectedService}/${currentMethod.name}`;

        const fetchResponse = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
        });

        const endTime = performance.now();
        setResponseTime(Math.round(endTime - startTime));

        if (fetchResponse.ok) {
          const responseData = await fetchResponse.json();
          setResponse(responseData);
        } else {
          // Try to parse error response
          try {
            const errorData = await fetchResponse.json();
            setError(
              errorData.message ||
                errorData.error ||
                `HTTP ${fetchResponse.status}: ${fetchResponse.statusText}`
            );
          } catch {
            setError(`HTTP ${fetchResponse.status}: ${fetchResponse.statusText}`);
          }
        }
      } else {
        // gRPC mode - use backend proxy
        const result = await catalogClient.invokeGRPC({
          endpoint: targetEndpoint,
          service: selectedService,
          method: currentMethod.name,
          requestJson: JSON.stringify(request),
          useTls: useTls,
          timeoutSeconds: 30,
          transport: transport,
        });

        const endTime = performance.now();
        setResponseTime(Math.round(endTime - startTime));

        // Check if the invocation was successful
        if (result.success) {
          if (result.responseJson) {
            setResponse(JSON.parse(result.responseJson));
          } else {
            setResponse({}); // Empty successful response
          }
        } else {
          // Invocation failed - display error from the response
          const errorMsg = result.error || result.statusMessage || 'Request failed';
          setError(errorMsg);
        }
      }
    } catch (err) {
      console.error('Request failed:', err);
      setError(err instanceof Error ? err.message : 'Request failed');
    } finally {
      setLoading(false);
    }
  };

  const clearResponse = () => {
    setResponse(undefined);
    setResponseTime(undefined);
    setError(undefined);
  };

  return {
    response,
    responseTime,
    error,
    loading,
    invokeMethod,
    clearResponse,
  };
}
