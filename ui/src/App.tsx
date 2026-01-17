import { useState, useEffect } from 'react';
import { ServiceBrowser } from '@/components/ServiceBrowser';
import { MethodDetails } from '@/components/MethodDetails';
import { RequestEditor } from '@/components/RequestEditor';
import { ResponseViewer } from '@/components/ResponseViewer';
import { LoadProtos } from '@/components/LoadProtos';
import { catalogClient, clearSession } from '@/lib/client';
import { Transport } from '@gen/catalog/v1/catalog_pb';
import type { ServiceInfo, MethodInfo, MessageSchema } from '@/lib/types';

function App() {
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [selectedMethod, setSelectedMethod] = useState<string | null>(null);
  const [currentMethod, setCurrentMethod] = useState<MethodInfo | null>(null);
  const [inputSchema, setInputSchema] = useState<MessageSchema | undefined>();
  const [outputSchema, setOutputSchema] = useState<MessageSchema | undefined>();
  const [response, setResponse] = useState<Record<string, unknown> | undefined>();
  const [responseTime, setResponseTime] = useState<number | undefined>();
  const [error, setError] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);
  const [initializing, setInitializing] = useState(true);
  const [targetEndpoint, setTargetEndpoint] = useState('localhost:50051');
  const [useTls, setUseTls] = useState(false);
  const [transport, setTransport] = useState<Transport>(Transport.CONNECT);

  // Auto-detect TLS based on port
  const handleEndpointChange = (newEndpoint: string) => {
    setTargetEndpoint(newEndpoint);
    // Auto-enable TLS for port 443
    if (newEndpoint.includes(':443')) {
      setUseTls(true);
    }
  };

  useEffect(() => {
    loadServices();
  }, []);

  const loadServices = async (newEndpoint?: string) => {
    try {
      setInitializing(true);
      const result = await catalogClient.listServices({});

      // Get current service names to identify new ones
      const existingServiceNames = new Set(services.map(s => s.fullName));

      const servicesData: ServiceInfo[] = ((result as any).services || []).map((svc: any) => {
        const isNewService = !existingServiceNames.has(svc.name);
        // Find existing service to preserve its endpoint, or use new endpoint for new services
        const existingService = services.find(s => s.fullName === svc.name);

        return {
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
          // Preserve existing endpoint or assign new endpoint to newly loaded services
          endpoint: existingService?.endpoint || (isNewService ? newEndpoint : undefined),
        };
      });

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
    setResponseTime(undefined);
    setError(undefined);

    const service = services.find((s) => s.fullName === serviceFullName);
    const method = service?.methods.find((m) => m.name === methodName);

    // Auto-update target endpoint if service has a known endpoint
    if (service?.endpoint) {
      handleEndpointChange(service.endpoint);
    }

    if (method) {
      setCurrentMethod(method);

      try {
        const schemaResult = await catalogClient.getServiceSchema({
          serviceName: serviceFullName,
        });

        // Get method info from service.methods
        const methodInfo = (schemaResult as any).service?.methods
          ?.find((m: any) => m.name === methodName);

        // Get message schemas from messageSchemas map (they're JSON strings)
        const messageSchemas = (schemaResult as any).messageSchemas || {};

        if (methodInfo) {
          // Parse input schema from JSON string
          const inputSchemaJson = messageSchemas[methodInfo.inputType];
          if (inputSchemaJson) {
            try {
              const parsedInput = JSON.parse(inputSchemaJson);
              setInputSchema({
                name: parsedInput.title || methodInfo.inputType.split('.').pop() || '',
                fields: Object.entries(parsedInput.properties || {}).map(([name, prop]: [string, any]) => ({
                  name,
                  type: prop.type || 'string',
                  repeated: prop.type === 'array',
                  optional: !(parsedInput.required || []).includes(name),
                  description: prop.description,
                })),
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
                fields: Object.entries(parsedOutput.properties || {}).map(([name, prop]: [string, any]) => ({
                  name,
                  type: prop.type || 'string',
                  repeated: prop.type === 'array',
                  optional: !(parsedOutput.required || []).includes(name),
                  description: prop.description,
                })),
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
    }
  };

  const handleSubmitRequest = async (request: Record<string, unknown>) => {
    if (!selectedService || !currentMethod) return;

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
            setError(errorData.message || errorData.error || `HTTP ${fetchResponse.status}: ${fetchResponse.statusText}`);
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
        if ((result as any).success) {
          if ((result as any).responseJson) {
            setResponse(JSON.parse((result as any).responseJson));
          } else {
            setResponse({}); // Empty successful response
          }
        } else {
          // Invocation failed - display error from the response
          const errorMsg = (result as any).error || (result as any).statusMessage || 'Request failed';
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

  if (initializing) {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent mx-auto" />
          <p className="text-sm text-muted-foreground">Loading services...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen flex flex-col bg-background">
      {/* Header */}
      <header className="h-14 border-b bg-card/50 backdrop-blur-sm flex items-center justify-between px-4 shrink-0">
        <div className="flex items-center gap-3">
          <button
            onClick={() => {
              setSelectedService(null);
              setSelectedMethod(null);
              setCurrentMethod(null);
              setServices([]);
              setResponse(undefined);
              setError(undefined);
              clearSession(); // Clear session to start fresh
            }}
            className="flex items-center gap-2 hover:opacity-80 transition-opacity cursor-pointer"
            title="Load new protos"
          >
            <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
              <svg className="w-5 h-5 text-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <span className="font-semibold text-lg">ConnectRPC</span>
            <span className="text-muted-foreground font-normal">Inspect</span>
          </button>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Target:</span>
            <input
              type="text"
              value={targetEndpoint}
              onChange={(e) => handleEndpointChange(e.target.value)}
              placeholder="localhost:50051"
              className="h-8 w-48 px-3 text-sm rounded-md border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
            <label className="flex items-center gap-1.5 cursor-pointer">
              <input
                type="checkbox"
                checked={useTls}
                onChange={(e) => setUseTls(e.target.checked)}
                className="rounded border-border h-4 w-4"
              />
              <span className="text-sm text-muted-foreground">TLS</span>
            </label>
          </div>
          <div className="flex items-center h-8 rounded-md border bg-muted/50 p-0.5">
            <button
              onClick={() => setTransport(Transport.CONNECT)}
              className={`px-3 h-7 text-sm rounded-[5px] transition-colors ${
                transport === Transport.CONNECT
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              HTTP
            </button>
            <button
              onClick={() => setTransport(Transport.GRPC)}
              className={`px-3 h-7 text-sm rounded-[5px] transition-colors ${
                transport === Transport.GRPC
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              gRPC
            </button>
          </div>
        </div>
      </header>

      {/* Main Content - Three Column Layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left Column: Service Browser */}
        <aside className="w-64 border-r bg-card/30 flex flex-col shrink-0">
          <ServiceBrowser
            services={services}
            selectedMethod={selectedMethod}
            onSelectMethod={handleSelectMethod}
          />
        </aside>

        {/* Center Column: Method Documentation */}
        <main className="flex-1 overflow-auto border-r bg-background">
          {!currentMethod ? (
            <div className="flex items-center justify-center h-full p-8">
              {services.length === 0 ? (
                <div className="w-full max-w-lg">
                  <LoadProtos onLoadSuccess={loadServices} />
                </div>
              ) : (
                <div className="text-center space-y-3 max-w-md">
                  <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center mx-auto">
                    <svg className="w-6 h-6 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                      <polyline points="14 2 14 8 20 8" />
                      <line x1="16" y1="13" x2="8" y2="13" />
                      <line x1="16" y1="17" x2="8" y2="17" />
                      <polyline points="10 9 9 9 8 9" />
                    </svg>
                  </div>
                  <div>
                    <p className="text-lg font-medium">Select a method</p>
                    <p className="text-sm text-muted-foreground mt-1">
                      Choose a service and method from the sidebar to view its documentation and try it out
                    </p>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="p-6 max-w-2xl">
              <MethodDetails
                method={currentMethod}
                inputSchema={inputSchema}
                outputSchema={outputSchema}
              />
            </div>
          )}
        </main>

        {/* Right Column: Interactive Playground */}
        <aside className="w-[420px] bg-muted/30 flex flex-col shrink-0 overflow-hidden">
          {currentMethod ? (
            <div className="flex flex-col h-full">
              {/* Request Section */}
              <div className="flex-1 overflow-auto p-4 border-b">
                <RequestEditor
                  schema={inputSchema}
                  onSubmit={handleSubmitRequest}
                  loading={loading}
                />
              </div>

              {/* Response Section */}
              <div className="flex-1 overflow-auto p-4">
                <ResponseViewer
                  response={response}
                  error={error}
                  loading={loading}
                  responseTime={responseTime}
                />
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center h-full p-8">
              <div className="text-center space-y-2">
                <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center mx-auto">
                  <svg className="w-5 h-5 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polygon points="5 3 19 12 5 21 5 3" />
                  </svg>
                </div>
                <p className="text-sm text-muted-foreground">
                  Select a method to try it out
                </p>
              </div>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}

export default App;
