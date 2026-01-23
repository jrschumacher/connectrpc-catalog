import { useState } from 'react';
import { catalogClient } from '@/lib/client';
import type { ServiceInfo, MethodInfo, MessageSchema } from '@/lib/types';

interface UseMethodSchemaReturn {
  selectedService: string | null;
  selectedMethod: string | null;
  currentMethod: MethodInfo | null;
  inputSchema: MessageSchema | undefined;
  outputSchema: MessageSchema | undefined;
  error: string | undefined;
  selectMethod: (serviceFullName: string, methodName: string) => Promise<void>;
  clearSelection: () => void;
}

/**
 * Hook for managing method selection and schema loading
 */
export function useMethodSchema(services: ServiceInfo[]): UseMethodSchemaReturn {
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [selectedMethod, setSelectedMethod] = useState<string | null>(null);
  const [currentMethod, setCurrentMethod] = useState<MethodInfo | null>(null);
  const [inputSchema, setInputSchema] = useState<MessageSchema | undefined>();
  const [outputSchema, setOutputSchema] = useState<MessageSchema | undefined>();
  const [error, setError] = useState<string | undefined>();

  const selectMethod = async (serviceFullName: string, methodName: string) => {
    setSelectedService(serviceFullName);
    setSelectedMethod(`${serviceFullName}.${methodName}`);
    setError(undefined);

    const service = services.find((s) => s.fullName === serviceFullName);
    const method = service?.methods.find((m) => m.name === methodName);

    if (!method) {
      setError('Method not found');
      return;
    }

    setCurrentMethod(method);

    try {
      const schemaResult = await catalogClient.getServiceSchema({
        serviceName: serviceFullName,
      });

      // Get method info from service.methods
      const methodInfo = schemaResult.service?.methods?.find(
        (m) => m.name === methodName
      );

      // Get message schemas from messageSchemas map (they're JSON strings)
      const messageSchemas = schemaResult.messageSchemas || {};

      if (methodInfo) {
        // Parse input schema from JSON string
        const inputSchemaJson = messageSchemas[methodInfo.inputType];
        if (inputSchemaJson) {
          try {
            const parsedInput = JSON.parse(inputSchemaJson);
            setInputSchema({
              name: parsedInput.title || methodInfo.inputType.split('.').pop() || '',
              fields: Object.entries(parsedInput.properties || {}).map(
                ([name, prop]: [string, any]) => ({
                  name,
                  type: prop.type || 'string',
                  repeated: prop.type === 'array',
                  optional: !(parsedInput.required || []).includes(name),
                  description: prop.description,
                })
              ),
            });
          } catch (e) {
            console.error('Failed to parse input schema JSON:', e);
          }
        }

        // Parse output schema from JSON string
        const outputSchemaJson = messageSchemas[methodInfo.outputType];
        if (outputSchemaJson) {
          try {
            const parsedOutput = JSON.parse(outputSchemaJson);
            setOutputSchema({
              name: parsedOutput.title || methodInfo.outputType.split('.').pop() || '',
              fields: Object.entries(parsedOutput.properties || {}).map(
                ([name, prop]: [string, any]) => ({
                  name,
                  type: prop.type || 'string',
                  repeated: prop.type === 'array',
                  optional: !(parsedOutput.required || []).includes(name),
                  description: prop.description,
                })
              ),
            });
          } catch (e) {
            console.error('Failed to parse output schema JSON:', e);
          }
        }
      }
    } catch (err) {
      console.error('Failed to load schema:', err);
      setError(err instanceof Error ? err.message : 'Failed to load schema');
    }
  };

  const clearSelection = () => {
    setSelectedService(null);
    setSelectedMethod(null);
    setCurrentMethod(null);
    setInputSchema(undefined);
    setOutputSchema(undefined);
    setError(undefined);
  };

  return {
    selectedService,
    selectedMethod,
    currentMethod,
    inputSchema,
    outputSchema,
    error,
    selectMethod,
    clearSelection,
  };
}
