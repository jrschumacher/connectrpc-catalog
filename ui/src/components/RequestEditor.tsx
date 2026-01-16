import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Send, Code, FormInput } from 'lucide-react';
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
      <div className="h-full flex items-center justify-center bg-card rounded-lg border">
        <div className="text-center space-y-2 p-6">
          <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center mx-auto">
            <FormInput className="w-5 h-5 text-muted-foreground" />
          </div>
          <p className="text-sm text-muted-foreground">
            Select a method to begin
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-medium text-sm">Request</h3>
        <div className="flex items-center gap-1 bg-muted rounded-md p-0.5">
          <button
            onClick={() => setJsonMode(false)}
            className={`flex items-center gap-1.5 px-2.5 py-1 text-xs rounded transition-colors ${
              !jsonMode
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <FormInput className="h-3 w-3" />
            Form
          </button>
          <button
            onClick={() => setJsonMode(true)}
            className={`flex items-center gap-1.5 px-2.5 py-1 text-xs rounded transition-colors ${
              jsonMode
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <Code className="h-3 w-3" />
            JSON
          </button>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleFormSubmit} className="flex-1 flex flex-col min-h-0">
        <div className="flex-1 overflow-auto">
          {jsonMode ? (
            <textarea
              value={jsonInput}
              onChange={(e) => setJsonInput(e.target.value)}
              className="w-full h-full min-h-[200px] rounded-lg border bg-card px-4 py-3 text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background"
              placeholder="{}"
              spellCheck={false}
            />
          ) : (
            <div className="space-y-4 bg-card rounded-lg border p-4">
              {schema.fields.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-4">
                  No input fields required
                </p>
              ) : (
                schema.fields.map((field) => (
                  <div key={field.name} className="space-y-1.5">
                    <Label htmlFor={field.name} className="text-sm font-medium">
                      {field.name}
                      {!field.optional && (
                        <span className="text-red-500 ml-0.5">*</span>
                      )}
                    </Label>
                    <Input
                      id={field.name}
                      value={formData[field.name] || ''}
                      onChange={(e) => handleFieldChange(field.name, e.target.value)}
                      placeholder={field.type + (field.repeated ? '[]' : '')}
                      className="font-mono text-sm"
                    />
                    {field.description && (
                      <p className="text-xs text-muted-foreground">{field.description}</p>
                    )}
                  </div>
                ))
              )}
            </div>
          )}
        </div>

        {/* Submit Button */}
        <div className="pt-4 shrink-0">
          <Button
            type="submit"
            disabled={loading}
            className="w-full"
            size="default"
          >
            {loading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-2 border-primary-foreground border-t-transparent mr-2" />
                Sending...
              </>
            ) : (
              <>
                <Send className="mr-2 h-4 w-4" />
                Send Request
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
};
