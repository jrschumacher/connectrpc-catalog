import React from 'react';
import { Badge } from '@/components/ui/badge';
import { ArrowRight, Circle, Zap } from 'lucide-react';
import type { MethodInfo, MessageSchema } from '@/lib/types';

interface MethodDetailsProps {
  method: MethodInfo;
  inputSchema?: MessageSchema;
  outputSchema?: MessageSchema;
}

const getMethodType = (method: MethodInfo) => {
  if (method.clientStreaming && method.serverStreaming) return 'Bidirectional';
  if (method.clientStreaming) return 'Client Stream';
  if (method.serverStreaming) return 'Server Stream';
  return 'Unary';
};

const getMethodBadgeClass = (method: MethodInfo) => {
  if (method.clientStreaming || method.serverStreaming) {
    return 'bg-purple-500/10 text-purple-600 border-purple-500/20';
  }
  return 'bg-blue-500/10 text-blue-600 border-blue-500/20';
};

const SchemaFields: React.FC<{ schema: MessageSchema; title: string }> = ({ schema, title }) => (
  <div className="space-y-3">
    <div className="flex items-center gap-2">
      <h4 className="text-sm font-medium text-muted-foreground">{title}</h4>
      <code className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
        {schema.name}
      </code>
    </div>
    <div className="border rounded-lg divide-y">
      {schema.fields.length === 0 ? (
        <div className="px-4 py-3 text-sm text-muted-foreground italic">
          No fields
        </div>
      ) : (
        schema.fields.map((field) => (
          <div key={field.name} className="px-4 py-3 flex items-start gap-3">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <code className="text-sm font-semibold font-mono">{field.name}</code>
                {field.optional && (
                  <span className="text-xs text-muted-foreground">optional</span>
                )}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <code className="text-xs text-muted-foreground font-mono">
                  {field.repeated && 'repeated '}
                  {field.type}
                </code>
              </div>
              {field.description && (
                <p className="text-xs text-muted-foreground mt-1.5">
                  {field.description}
                </p>
              )}
            </div>
          </div>
        ))
      )}
    </div>
  </div>
);

export const MethodDetails: React.FC<MethodDetailsProps> = ({
  method,
  inputSchema,
  outputSchema,
}) => {
  const methodType = getMethodType(method);

  return (
    <div className="space-y-8">
      {/* Method Header */}
      <div className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h2 className="text-2xl font-semibold tracking-tight">{method.name}</h2>
            <p className="text-sm text-muted-foreground font-mono">{method.fullName}</p>
          </div>
          <Badge variant="outline" className={getMethodBadgeClass(method)}>
            {methodType === 'Unary' ? (
              <Circle className="h-3 w-3 mr-1.5 fill-current" />
            ) : (
              <Zap className="h-3 w-3 mr-1.5" />
            )}
            {methodType}
          </Badge>
        </div>

        {/* HTTP Path */}
        <div className="flex items-center gap-2 text-sm">
          <Badge variant="secondary" className="font-mono text-xs">POST</Badge>
          <code className="text-muted-foreground font-mono">
            /{method.fullName.replace(/\./g, '/')}
          </code>
        </div>
      </div>

      {/* Type Flow */}
      <div className="flex items-center gap-3 p-4 bg-muted/50 rounded-lg">
        <div className="flex-1">
          <div className="text-xs text-muted-foreground mb-1">Request</div>
          <code className="text-sm font-mono font-medium">{method.inputType.split('.').pop()}</code>
        </div>
        <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0" />
        <div className="flex-1 text-right">
          <div className="text-xs text-muted-foreground mb-1">Response</div>
          <code className="text-sm font-mono font-medium">{method.outputType.split('.').pop()}</code>
        </div>
      </div>

      {/* Schemas */}
      {inputSchema && (
        <SchemaFields schema={inputSchema} title="Request Schema" />
      )}

      {outputSchema && (
        <SchemaFields schema={outputSchema} title="Response Schema" />
      )}
    </div>
  );
};
