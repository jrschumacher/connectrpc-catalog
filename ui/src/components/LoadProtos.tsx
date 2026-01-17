import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { catalogClient } from '@/lib/client';
import { Upload, Github, Package, Folder, CheckCircle, AlertCircle, Loader2, Zap } from 'lucide-react';

type SourceType = 'buf_module' | 'proto_path' | 'proto_repo' | 'reflection';

interface LoadProtosProps {
  onLoadSuccess: (endpoint?: string) => void;
}

const sourceOptions = [
  {
    type: 'buf_module' as SourceType,
    label: 'Buf Registry',
    icon: Package,
    placeholder: 'buf.build/connectrpc/eliza',
    description: 'Load from Buf Schema Registry',
  },
  {
    type: 'proto_path' as SourceType,
    label: 'Local Path',
    icon: Folder,
    placeholder: '/path/to/protos',
    description: 'Load from local filesystem',
  },
  {
    type: 'proto_repo' as SourceType,
    label: 'GitHub',
    icon: Github,
    placeholder: 'github.com/connectrpc/eliza',
    description: 'Load from GitHub repository',
  },
  {
    type: 'reflection' as SourceType,
    label: 'Auto-Discover',
    icon: Zap,
    placeholder: 'demo.connectrpc.com:443',
    description: 'Discover services via gRPC reflection',
  },
];

export function LoadProtos({ onLoadSuccess }: LoadProtosProps) {
  const [sourceType, setSourceType] = useState<SourceType>('buf_module');
  const [sourceValue, setSourceValue] = useState('');
  const [endpoint, setEndpoint] = useState('localhost:50051');
  const [useTls, setUseTls] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [success, setSuccess] = useState<string | undefined>();

  const handleLoad = async () => {
    if (!sourceValue.trim()) {
      setError('Please enter a source value');
      return;
    }

    setLoading(true);
    setError(undefined);
    setSuccess(undefined);

    try {
      // Build request with proper protobuf-es oneof format
      const request: any = {};

      if (sourceType === 'buf_module') {
        request.source = { case: 'bufModule', value: sourceValue };
      } else if (sourceType === 'proto_path') {
        request.source = { case: 'protoPath', value: sourceValue };
      } else if (sourceType === 'proto_repo') {
        request.source = { case: 'protoRepo', value: sourceValue };
      } else if (sourceType === 'reflection') {
        request.source = { case: 'reflectionEndpoint', value: sourceValue };
        request.reflectionOptions = {
          useTls: useTls,
          timeoutSeconds: 15,
        };
      }

      const result = await catalogClient.loadProtos(request);

      if ((result as any).success) {
        const serviceCount = (result as any).serviceCount || 0;
        const fileCount = (result as any).fileCount || 0;
        setSuccess(`Loaded ${serviceCount} service${serviceCount !== 1 ? 's' : ''} from ${fileCount} file${fileCount !== 1 ? 's' : ''}`);

        // For reflection, use the reflection endpoint as the target
        // For other sources, use the manually specified endpoint
        const targetEndpoint = sourceType === 'reflection'
          ? sourceValue.trim()
          : endpoint.trim();

        setSourceValue('');

        setTimeout(() => {
          onLoadSuccess(targetEndpoint || undefined);
        }, 500);
      } else {
        setError((result as any).error || 'Failed to load protos');
      }
    } catch (err) {
      console.error('Failed to load protos:', err);
      setError(err instanceof Error ? err.message : 'Failed to load protos');
    } finally {
      setLoading(false);
    }
  };

  const currentSource = sourceOptions.find((s) => s.type === sourceType)!;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto">
          <Upload className="w-7 h-7 text-primary" />
        </div>
        <h2 className="text-xl font-semibold">Load Proto Definitions</h2>
        <p className="text-sm text-muted-foreground">
          Import service definitions to explore and test APIs
        </p>
      </div>

      {/* Source Type Selection */}
      <div className="space-y-2">
        <Label className="text-sm font-medium">Source</Label>
        <div className="grid grid-cols-2 gap-2">
          {sourceOptions.map((option) => {
            const Icon = option.icon;
            const isSelected = sourceType === option.type;
            return (
              <button
                key={option.type}
                onClick={() => setSourceType(option.type)}
                className={`flex flex-col items-center gap-2 p-3 rounded-lg border transition-all ${
                  isSelected
                    ? 'bg-primary/5 border-primary text-primary'
                    : 'bg-card border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
                }`}
              >
                <Icon className="w-5 h-5" />
                <span className="text-xs font-medium">{option.label}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Source Input */}
      <div className="space-y-2">
        <Label htmlFor="source-input" className="text-sm font-medium">
          {currentSource.label}
        </Label>
        <Input
          id="source-input"
          placeholder={currentSource.placeholder}
          value={sourceValue}
          onChange={(e) => setSourceValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !loading) {
              handleLoad();
            }
          }}
          disabled={loading}
          className="font-mono text-sm"
        />
        <p className="text-xs text-muted-foreground">{currentSource.description}</p>

        {/* TLS checkbox for reflection mode */}
        {sourceType === 'reflection' && (
          <div className="flex items-center gap-2 pt-1">
            <input
              type="checkbox"
              id="use-tls"
              checked={useTls}
              onChange={(e) => setUseTls(e.target.checked)}
              className="rounded border-border"
            />
            <Label htmlFor="use-tls" className="text-sm cursor-pointer">Use TLS (required for https endpoints)</Label>
          </div>
        )}
      </div>

      {/* Target Endpoint - only show for non-reflection sources */}
      {sourceType !== 'reflection' && (
        <div className="space-y-2">
          <Label htmlFor="endpoint-input" className="text-sm font-medium">
            Target Endpoint
          </Label>
          <Input
            id="endpoint-input"
            placeholder="localhost:50051"
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            disabled={loading}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">Where this service runs (used when invoking methods)</p>
        </div>
      )}

      {/* Load Button */}
      <Button onClick={handleLoad} disabled={loading || !sourceValue.trim()} className="w-full">
        {loading ? (
          <Loader2 className="w-4 h-4 animate-spin mr-2" />
        ) : null}
        {loading ? 'Loading...' : 'Load Protos'}
      </Button>

      {/* Status Messages */}
      {error && (
        <div className="flex items-start gap-3 p-3 rounded-lg bg-red-500/5 border border-red-500/20">
          <AlertCircle className="w-5 h-5 text-red-500 shrink-0 mt-0.5" />
          <div className="space-y-1">
            <p className="text-sm font-medium text-red-600">Failed to load</p>
            <p className="text-sm text-red-600/80">{error}</p>
          </div>
        </div>
      )}

      {success && (
        <div className="flex items-start gap-3 p-3 rounded-lg bg-green-500/5 border border-green-500/20">
          <CheckCircle className="w-5 h-5 text-green-500 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-medium text-green-600">{success}</p>
          </div>
        </div>
      )}

      {/* Examples */}
      <div className="pt-4 border-t space-y-2">
        <p className="text-xs font-medium text-muted-foreground">Quick examples</p>
        <div className="grid gap-1.5">
          {[
            { type: 'buf_module', value: 'buf.build/connectrpc/eliza' },
            { type: 'buf_module', value: 'buf.build/grpc/grpc' },
            { type: 'reflection', value: 'demo.connectrpc.com:443' },
          ].map((example) => (
            <button
              key={example.value}
              onClick={() => {
                setSourceType(example.type as SourceType);
                setSourceValue(example.value);
              }}
              className="text-left text-xs font-mono px-2 py-1.5 rounded bg-muted hover:bg-accent transition-colors"
            >
              {example.value}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
