import React from 'react';
import { AlertCircle, CheckCircle, Clock, Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface ResponseViewerProps {
  response?: Record<string, unknown>;
  error?: string;
  loading?: boolean;
  responseTime?: number;
}

export const ResponseViewer: React.FC<ResponseViewerProps> = ({
  response,
  error,
  loading = false,
  responseTime,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (response) {
      await navigator.clipboard.writeText(JSON.stringify(response, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <h3 className="font-medium text-sm">Response</h3>
          {response && (
            <span className="flex items-center gap-1.5 text-xs text-green-600 bg-green-500/10 px-2 py-0.5 rounded-full">
              <CheckCircle className="h-3 w-3" />
              Success
            </span>
          )}
          {error && (
            <span className="flex items-center gap-1.5 text-xs text-red-600 bg-red-500/10 px-2 py-0.5 rounded-full">
              <AlertCircle className="h-3 w-3" />
              Error
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {responseTime !== undefined && (
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              {responseTime}ms
            </span>
          )}
          {response && (
            <button
              onClick={handleCopy}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {copied ? (
                <>
                  <Check className="h-3 w-3" />
                  Copied
                </>
              ) : (
                <>
                  <Copy className="h-3 w-3" />
                  Copy
                </>
              )}
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-h-0">
        {loading && (
          <div className="h-full flex items-center justify-center bg-card rounded-lg border">
            <div className="text-center space-y-3">
              <div className="animate-spin rounded-full h-6 w-6 border-2 border-primary border-t-transparent mx-auto" />
              <p className="text-xs text-muted-foreground">Sending request...</p>
            </div>
          </div>
        )}

        {!loading && !response && !error && (
          <div className="h-full flex items-center justify-center bg-card rounded-lg border">
            <div className="text-center space-y-2 p-6">
              <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center mx-auto">
                <svg className="w-5 h-5 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                </svg>
              </div>
              <p className="text-sm text-muted-foreground">
                Response will appear here
              </p>
            </div>
          </div>
        )}

        {error && (
          <div className="h-full overflow-auto bg-red-500/5 rounded-lg border border-red-500/20 p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="text-sm font-medium text-red-600">Request Failed</p>
                <p className="text-sm text-red-600/80 font-mono">{error}</p>
              </div>
            </div>
          </div>
        )}

        {response && (
          <div className="h-full overflow-auto bg-card rounded-lg border">
            <pre className="p-4 text-sm font-mono overflow-x-auto">
              <code className="text-foreground">{JSON.stringify(response, null, 2)}</code>
            </pre>
          </div>
        )}
      </div>
    </div>
  );
};
