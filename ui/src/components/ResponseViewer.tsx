import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { AlertCircle, CheckCircle } from 'lucide-react';

interface ResponseViewerProps {
  response?: Record<string, unknown>;
  error?: string;
  loading?: boolean;
}

export const ResponseViewer: React.FC<ResponseViewerProps> = ({
  response,
  error,
  loading = false,
}) => {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Response</CardTitle>
          {response && (
            <Badge variant="secondary" className="gap-1">
              <CheckCircle className="h-3 w-3" />
              Success
            </Badge>
          )}
          {error && (
            <Badge variant="destructive" className="gap-1">
              <AlertCircle className="h-3 w-3" />
              Error
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {loading && (
          <div className="flex items-center justify-center p-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        )}

        {!loading && !response && !error && (
          <div className="p-8 text-center text-sm text-muted-foreground">
            No response yet
          </div>
        )}

        {error && (
          <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive">
            {error}
          </div>
        )}

        {response && (
          <pre className="overflow-auto rounded-md bg-muted p-4 text-sm">
            <code>{JSON.stringify(response, null, 2)}</code>
          </pre>
        )}
      </CardContent>
    </Card>
  );
};
