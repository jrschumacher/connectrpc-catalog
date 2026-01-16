import { useState, useEffect } from 'react';
import { ServiceBrowser } from '@/components/ServiceBrowser';
import { MethodDetails } from '@/components/MethodDetails';
import { RequestEditor } from '@/components/RequestEditor';
import { ResponseViewer } from '@/components/ResponseViewer';
import { LoadProtos } from '@/components/LoadProtos';
import { catalogClient } from '@/lib/client';
import type { ServiceInfo, MethodInfo, MessageSchema } from '@/lib/types';

function App() {
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [selectedMethod, setSelectedMethod] = useState<string | null>(null);
  const [currentMethod, setCurrentMethod] = useState<MethodInfo | null>(null);
  const [inputSchema, setInputSchema] = useState<MessageSchema | undefined>();
  const [outputSchema, setOutputSchema] = useState<MessageSchema | undefined>();
  const [response, setResponse] = useState<Record<string, unknown> | undefined>();
  const [error, setError] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);
  const [initializing, setInitializing] = useState(true);

  useEffect(() => {
    loadServices();
  }, []);

  const loadServices = async () => {
    try {
      setInitializing(true);
      const result = await catalogClient.listServices({});

      const servicesData: ServiceInfo[] = ((result as any).services || []).map((svc: any) => ({
        name: svc.name || '',
        fullName: svc.name || '',
        methods: (svc.methods || []).map((method: any) => ({
          name: method.name || '',
          fullName: `${svc.name}.${method.name}`,
          inputType: method.inputType || '',
          outputType: method.outputType || '',
          clientStreaming: method.clientStreaming || false,
          serverStreaming: method.serverStreaming || false,
        })),
      }));

      setServices(servicesData);
    } catch (err) {
      console.error('Failed to load services:', err);
      setError(err instanceof Error ? err.message : 'Failed to load services');
    } finally {
      setInitializing(false);
    }
  };

  const handleSelectMethod = async (serviceFullName: string, methodName: string) => {
    setSelectedService(serviceFullName);
    setSelectedMethod(`${serviceFullName}.${methodName}`);
    setResponse(undefined);
    setError(undefined);

    const service = services.find((s) => s.fullName === serviceFullName);
    const method = service?.methods.find((m) => m.name === methodName);

    if (method) {
      setCurrentMethod(method);

      try {
        const schemaResult = await catalogClient.getServiceSchema({
          serviceName: serviceFullName,
        });

        const inputMsg = (schemaResult as any).methods
          ?.find((m: any) => m.name === methodName)
          ?.inputSchema;

        const outputMsg = (schemaResult as any).methods
          ?.find((m: any) => m.name === methodName)
          ?.outputSchema;

        if (inputMsg) {
          setInputSchema({
            name: inputMsg.name || '',
            fields: (inputMsg.fields || []).map((f: any) => ({
              name: f.name || '',
              type: f.type || '',
              repeated: f.repeated || false,
              optional: f.optional || false,
              description: f.description,
            })),
          });
        }

        if (outputMsg) {
          setOutputSchema({
            name: outputMsg.name || '',
            fields: (outputMsg.fields || []).map((f: any) => ({
              name: f.name || '',
              type: f.type || '',
              repeated: f.repeated || false,
              optional: f.optional || false,
              description: f.description,
            })),
          });
        }
      } catch (err) {
        console.error('Failed to load schema:', err);
        setError(err instanceof Error ? err.message : 'Failed to load schema');
      }
    }
  };

  const handleSubmitRequest = async (request: Record<string, unknown>) => {
    if (!selectedService || !currentMethod) return;

    setLoading(true);
    setError(undefined);
    setResponse(undefined);

    try {
      const result = await catalogClient.invokeGRPC({
        endpoint: 'localhost:8080',
        service: selectedService,
        method: currentMethod.name,
        requestJson: JSON.stringify(request),
        useTls: false,
        timeoutSeconds: 30,
      });

      if ((result as any).responseJson) {
        setResponse(JSON.parse((result as any).responseJson));
      }
    } catch (err) {
      console.error('Request failed:', err);
      setError(err instanceof Error ? err.message : 'Request failed');
    } finally {
      setLoading(false);
    }
  };

  if (initializing) {
    return (
      <div className="h-screen w-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-sm text-muted-foreground">Loading services...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen flex flex-col">
      <header className="border-b px-6 py-4">
        <h1 className="text-2xl font-bold">ConnectRPC Catalog</h1>
        <p className="text-sm text-muted-foreground">
          Browse and test gRPC services
        </p>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <aside className="w-80 border-r flex flex-col">
          <ServiceBrowser
            services={services}
            selectedMethod={selectedMethod}
            onSelectMethod={handleSelectMethod}
          />
        </aside>

        <main className="flex-1 overflow-auto">
          {!currentMethod ? (
            <div className="flex items-center justify-center h-full p-6">
              {services.length === 0 ? (
                <div className="w-full max-w-2xl">
                  <LoadProtos onLoadSuccess={loadServices} />
                </div>
              ) : (
                <div className="text-center space-y-2">
                  <p className="text-lg font-medium">No method selected</p>
                  <p className="text-sm text-muted-foreground">
                    Select a service and method from the sidebar to get started
                  </p>
                </div>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-6 p-6 h-full">
              <div className="space-y-6 overflow-auto">
                <MethodDetails
                  method={currentMethod}
                  inputSchema={inputSchema}
                  outputSchema={outputSchema}
                />
                <RequestEditor
                  schema={inputSchema}
                  onSubmit={handleSubmitRequest}
                  loading={loading}
                />
              </div>

              <div className="overflow-auto">
                <ResponseViewer
                  response={response}
                  error={error}
                  loading={loading}
                />
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

export default App;
