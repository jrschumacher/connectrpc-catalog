import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { catalogClient } from '@/lib/client';

type SourceType = 'buf_module' | 'proto_path' | 'proto_repo';

interface LoadProtosProps {
  onLoadSuccess: () => void;
}

export function LoadProtos({ onLoadSuccess }: LoadProtosProps) {
  const [sourceType, setSourceType] = useState<SourceType>('buf_module');
  const [sourceValue, setSourceValue] = useState('');
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
      const request: any = {};

      // Map source type to proto field name
      if (sourceType === 'buf_module') {
        request.bufModule = sourceValue;
      } else if (sourceType === 'proto_path') {
        request.protoPath = sourceValue;
      } else if (sourceType === 'proto_repo') {
        request.protoRepo = sourceValue;
      }

      const result = await catalogClient.loadProtos(request);

      if ((result as any).success) {
        const serviceCount = (result as any).serviceCount || 0;
        const fileCount = (result as any).fileCount || 0;
        setSuccess(`Successfully loaded ${serviceCount} services from ${fileCount} files`);
        setSourceValue('');

        // Notify parent to refresh services
        setTimeout(() => {
          onLoadSuccess();
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

  const getPlaceholder = () => {
    switch (sourceType) {
      case 'buf_module':
        return 'buf.build/connectrpc/eliza';
      case 'proto_path':
        return '/path/to/proto/files';
      case 'proto_repo':
        return 'github.com/connectrpc/eliza';
      default:
        return '';
    }
  };

  const getDescription = () => {
    switch (sourceType) {
      case 'buf_module':
        return 'Load from Buf Schema Registry (e.g., buf.build/owner/repo or owner/repo)';
      case 'proto_path':
        return 'Load from local filesystem directory containing .proto files';
      case 'proto_repo':
        return 'Load from GitHub repository (e.g., github.com/owner/repo)';
      default:
        return '';
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Load Proto Definitions</CardTitle>
        <CardDescription>
          Load service definitions from various sources to explore APIs
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label>Source Type</Label>
          <div className="flex gap-2">
            <Button
              variant={sourceType === 'buf_module' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSourceType('buf_module')}
            >
              Buf Module
            </Button>
            <Button
              variant={sourceType === 'proto_path' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSourceType('proto_path')}
            >
              Local Path
            </Button>
            <Button
              variant={sourceType === 'proto_repo' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSourceType('proto_repo')}
            >
              GitHub Repo
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">{getDescription()}</p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="source-value">Source</Label>
          <div className="flex gap-2">
            <Input
              id="source-value"
              placeholder={getPlaceholder()}
              value={sourceValue}
              onChange={(e) => setSourceValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !loading) {
                  handleLoad();
                }
              }}
              disabled={loading}
            />
            <Button onClick={handleLoad} disabled={loading || !sourceValue.trim()}>
              {loading ? 'Loading...' : 'Load'}
            </Button>
          </div>
        </div>

        {error && (
          <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
            <p className="text-sm text-destructive font-medium">Error</p>
            <p className="text-sm text-destructive/80 mt-1">{error}</p>
          </div>
        )}

        {success && (
          <div className="p-3 rounded-md bg-green-500/10 border border-green-500/20">
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="bg-green-500/20 text-green-700 border-green-500/30">
                Success
              </Badge>
              <p className="text-sm text-green-700">{success}</p>
            </div>
          </div>
        )}

        <div className="pt-2 border-t">
          <p className="text-xs text-muted-foreground">
            <strong>Examples:</strong>
          </p>
          <ul className="text-xs text-muted-foreground mt-1 space-y-1 ml-4">
            <li>• Buf Module: <code className="bg-muted px-1 py-0.5 rounded">buf.build/connectrpc/eliza</code></li>
            <li>• Local Path: <code className="bg-muted px-1 py-0.5 rounded">./proto</code></li>
            <li>• GitHub: <code className="bg-muted px-1 py-0.5 rounded">github.com/connectrpc/eliza</code></li>
          </ul>
        </div>
      </CardContent>
    </Card>
  );
}
