import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Send } from 'lucide-react';
import type { MessageSchema } from '@/lib/types';

interface RequestEditorProps {
  schema?: MessageSchema;
  onSubmit: (request: Record<string, unknown>) => void;
  loading?: boolean;
}

export const RequestEditor: React.FC<RequestEditorProps> = ({
  schema,
  onSubmit,
  loading = false,
}) => {
  const [jsonMode, setJsonMode] = useState(false);
  const [jsonInput, setJsonInput] = useState('{}');
  const [formData, setFormData] = useState<Record<string, string>>({});

  useEffect(() => {
    if (schema) {
      const initial: Record<string, string> = {};
      schema.fields.forEach((field) => {
        initial[field.name] = '';
      });
      setFormData(initial);
      setJsonInput(JSON.stringify({}, null, 2));
    }
  }, [schema]);

  const handleFormSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (jsonMode) {
      try {
        const parsed = JSON.parse(jsonInput);
        onSubmit(parsed);
      } catch (error) {
        alert('Invalid JSON: ' + (error instanceof Error ? error.message : 'Unknown error'));
      }
    } else {
      const request: Record<string, unknown> = {};
      Object.entries(formData).forEach(([key, value]) => {
        if (value !== '') {
          request[key] = value;
        }
      });
      onSubmit(request);
    }
  };

  const handleFieldChange = (fieldName: string, value: string) => {
    setFormData((prev) => ({ ...prev, [fieldName]: value }));
  };

  if (!schema) {
    return (
      <Card>
        <CardContent className="p-6">
          <p className="text-sm text-muted-foreground text-center">
            Select a method to begin
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Request</CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setJsonMode(!jsonMode)}
          >
            {jsonMode ? 'Form' : 'JSON'}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleFormSubmit} className="space-y-4">
          {jsonMode ? (
            <div>
              <Label htmlFor="json-input">JSON Request</Label>
              <textarea
                id="json-input"
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                className="min-h-[200px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="{}"
              />
            </div>
          ) : (
            <div className="space-y-3">
              {schema.fields.map((field) => (
                <div key={field.name} className="space-y-1">
                  <Label htmlFor={field.name}>
                    {field.name}
                    {!field.optional && (
                      <span className="text-destructive ml-1">*</span>
                    )}
                  </Label>
                  <Input
                    id={field.name}
                    value={formData[field.name] || ''}
                    onChange={(e) =>
                      handleFieldChange(field.name, e.target.value)
                    }
                    placeholder={`${field.type}${field.repeated ? '[]' : ''}`}
                    required={!field.optional}
                  />
                </div>
              ))}
            </div>
          )}

          <Button type="submit" disabled={loading} className="w-full">
            <Send className="mr-2 h-4 w-4" />
            {loading ? 'Sending...' : 'Send Request'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
};
