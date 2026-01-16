import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { MethodInfo, MessageSchema } from '@/lib/types';

interface MethodDetailsProps {
  method: MethodInfo;
  inputSchema?: MessageSchema;
  outputSchema?: MessageSchema;
}

export const MethodDetails: React.FC<MethodDetailsProps> = ({
  method,
  inputSchema,
  outputSchema,
}) => {
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{method.name}</CardTitle>
          <CardDescription className="font-mono text-xs">
            {method.fullName}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            {method.clientStreaming && (
              <Badge variant="outline">Client Streaming</Badge>
            )}
            {method.serverStreaming && (
              <Badge variant="outline">Server Streaming</Badge>
            )}
            {!method.clientStreaming && !method.serverStreaming && (
              <Badge variant="outline">Unary</Badge>
            )}
          </div>

          <div className="space-y-2">
            <div>
              <h4 className="text-sm font-medium mb-1">Input Type</h4>
              <code className="text-xs text-muted-foreground">
                {method.inputType}
              </code>
            </div>
            <div>
              <h4 className="text-sm font-medium mb-1">Output Type</h4>
              <code className="text-xs text-muted-foreground">
                {method.outputType}
              </code>
            </div>
          </div>
        </CardContent>
      </Card>

      {inputSchema && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Input Schema</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {inputSchema.fields.map((field) => (
                <div
                  key={field.name}
                  className="flex items-start gap-2 text-sm border-l-2 border-muted pl-3 py-1"
                >
                  <code className="font-mono font-medium">{field.name}</code>
                  <code className="text-muted-foreground">
                    {field.repeated && '[]'}
                    {field.type}
                  </code>
                  {field.optional && (
                    <Badge variant="secondary" className="text-xs">
                      optional
                    </Badge>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {outputSchema && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Output Schema</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {outputSchema.fields.map((field) => (
                <div
                  key={field.name}
                  className="flex items-start gap-2 text-sm border-l-2 border-muted pl-3 py-1"
                >
                  <code className="font-mono font-medium">{field.name}</code>
                  <code className="text-muted-foreground">
                    {field.repeated && '[]'}
                    {field.type}
                  </code>
                  {field.optional && (
                    <Badge variant="secondary" className="text-xs">
                      optional
                    </Badge>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};
