import { useState, useEffect } from 'react';
import { catalogClient } from '@/lib/client';
import type { ServiceInfo } from '@/lib/types';

interface UseServicesReturn {
  services: ServiceInfo[];
  initializing: boolean;
  error: string | undefined;
  loadServices: (newEndpoint?: string) => Promise<void>;
}

/**
 * Hook for managing service list state and loading
 */
export function useServices(): UseServicesReturn {
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [initializing, setInitializing] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const loadServices = async (newEndpoint?: string) => {
    try {
      setInitializing(true);
      const result = await catalogClient.listServices({});

      // Get current service names to identify new ones
      const existingServiceNames = new Set(services.map(s => s.fullName));

      const servicesData: ServiceInfo[] = (result.services || []).map((svc) => {
        const isNewService = !existingServiceNames.has(svc.name);
        // Find existing service to preserve its endpoint, or use new endpoint for new services
        const existingService = services.find(s => s.fullName === svc.name);

        return {
          name: svc.name || '',
          fullName: svc.name || '',
          methods: (svc.methods || []).map((method) => ({
            name: method.name || '',
            fullName: `${svc.name}.${method.name}`,
            inputType: method.inputType || '',
            outputType: method.outputType || '',
            clientStreaming: method.clientStreaming || false,
            serverStreaming: method.serverStreaming || false,
          })),
          // Preserve existing endpoint or assign new endpoint to newly loaded services
          endpoint: existingService?.endpoint || (isNewService ? newEndpoint : undefined),
        };
      });

      setServices(servicesData);
      setError(undefined);
    } catch (err) {
      console.error('Failed to load services:', err);
      setError(err instanceof Error ? err.message : 'Failed to load services');
    } finally {
      setInitializing(false);
    }
  };

  useEffect(() => {
    loadServices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    services,
    initializing,
    error,
    loadServices,
  };
}
