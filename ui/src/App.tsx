import { useState } from 'react';
import { ServiceBrowser } from '@/components/ServiceBrowser';
import { MethodDetails } from '@/components/MethodDetails';
import { RequestEditor } from '@/components/RequestEditor';
import { ResponseViewer } from '@/components/ResponseViewer';
import { LoadProtos } from '@/components/LoadProtos';
import { clearSession } from '@/lib/client';
import { Transport } from '@gen/catalog/v1/catalog_pb';
import { useServices } from '@/hooks/useServices';
import { useMethodSchema } from '@/hooks/useMethodSchema';
import { useMethodInvocation } from '@/hooks/useMethodInvocation';

function App() {
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

  // Custom hooks for state management
  const { services, initializing, loadServices } = useServices();

  const {
    selectedService,
    selectedMethod,
    currentMethod,
    inputSchema,
    outputSchema,
    selectMethod,
    clearSelection,
  } = useMethodSchema(services);

  const { response, responseTime, error, loading, invokeMethod, clearResponse } =
    useMethodInvocation({
      selectedService,
      currentMethod,
      targetEndpoint,
      useTls,
      transport,
    });

  const handleSelectMethod = async (serviceFullName: string, methodName: string) => {
    clearResponse();
    await selectMethod(serviceFullName, methodName);

    // Auto-update target endpoint if service has a known endpoint
    const service = services.find((s) => s.fullName === serviceFullName);
    if (service?.endpoint) {
      handleEndpointChange(service.endpoint);
    }
  };

  const handleReset = () => {
    clearSelection();
    clearResponse();
    clearSession();
    loadServices();
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
            onClick={handleReset}
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
                  onSubmit={invokeMethod}
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
