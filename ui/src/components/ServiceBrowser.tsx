import { useState, useEffect } from 'react';
import { ChevronRight, ChevronDown, Search, Circle, Zap, Package } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import type { ServiceInfo } from '@/lib/types';

interface ServiceBrowserProps {
  services: ServiceInfo[];
  selectedMethod: string | null;
  onSelectMethod: (serviceFullName: string, methodName: string) => void;
}

export const ServiceBrowser: React.FC<ServiceBrowserProps> = ({
  services,
  selectedMethod,
  onSelectMethod,
}) => {
  const [expandedServices, setExpandedServices] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');

  // Expand all services when they load
  useEffect(() => {
    if (services.length > 0) {
      setExpandedServices(new Set(services.map(s => s.fullName)));
    }
  }, [services]);

  const toggleService = (serviceName: string) => {
    const newExpanded = new Set(expandedServices);
    if (newExpanded.has(serviceName)) {
      newExpanded.delete(serviceName);
    } else {
      newExpanded.add(serviceName);
    }
    setExpandedServices(newExpanded);
  };

  const filteredServices = services.filter((service) => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      service.name.toLowerCase().includes(query) ||
      service.methods.some((m) => m.name.toLowerCase().includes(query))
    );
  });

  const getShortServiceName = (fullName: string) => {
    const parts = fullName.split('.');
    return parts[parts.length - 1];
  };

  const getPackageName = (fullName: string) => {
    const parts = fullName.split('.');
    return parts.slice(0, -1).join('.');
  };

  return (
    <div className="flex h-full flex-col">
      {/* Search */}
      <div className="p-3 border-b">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-9 text-sm bg-background"
          />
        </div>
      </div>

      {/* Services List */}
      <ScrollArea className="flex-1">
        <div className="p-2">
          {filteredServices.length === 0 ? (
            <div className="p-4 text-sm text-muted-foreground text-center">
              {services.length === 0 ? (
                <div className="space-y-2">
                  <Package className="h-8 w-8 mx-auto text-muted-foreground/50" />
                  <p>No services loaded</p>
                </div>
              ) : (
                'No matches found'
              )}
            </div>
          ) : (
            filteredServices.map((service) => {
              const isExpanded = expandedServices.has(service.fullName);
              const shortName = getShortServiceName(service.fullName);
              const packageName = getPackageName(service.fullName);

              return (
                <div key={service.fullName} className="mb-1">
                  {/* Service Header */}
                  <button
                    onClick={() => toggleService(service.fullName)}
                    className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm hover:bg-accent group"
                  >
                    <span className="text-muted-foreground">
                      {isExpanded ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                    </span>
                    <div className="flex-1 text-left min-w-0">
                      <div className="font-medium truncate">{shortName}</div>
                      {packageName && (
                        <div className="text-xs text-muted-foreground truncate">
                          {packageName}
                        </div>
                      )}
                      {service.endpoint && (
                        <div className="text-xs text-blue-500 truncate font-mono">
                          {service.endpoint}
                        </div>
                      )}
                    </div>
                    <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                      {service.methods.length}
                    </span>
                  </button>

                  {/* Methods List */}
                  {isExpanded && (
                    <div className="ml-4 mt-1 space-y-0.5 border-l pl-2">
                      {service.methods.map((method) => {
                        const methodId = `${service.fullName}.${method.name}`;
                        const isSelected = selectedMethod === methodId;
                        const isStreaming = method.clientStreaming || method.serverStreaming;

                        return (
                          <button
                            key={method.name}
                            onClick={() => onSelectMethod(service.fullName, method.name)}
                            className={cn(
                              'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
                              isSelected
                                ? 'bg-primary text-primary-foreground'
                                : 'hover:bg-accent text-foreground'
                            )}
                          >
                            {isStreaming ? (
                              <Zap className={cn(
                                'h-3 w-3 shrink-0',
                                isSelected ? 'text-primary-foreground' : 'text-purple-500'
                              )} />
                            ) : (
                              <Circle className={cn(
                                'h-3 w-3 shrink-0',
                                isSelected ? 'text-primary-foreground fill-current' : 'text-blue-500 fill-current'
                              )} />
                            )}
                            <span className="truncate">{method.name}</span>
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="p-3 border-t">
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{services.length} service{services.length !== 1 ? 's' : ''}</span>
          <span>{services.reduce((acc, s) => acc + s.methods.length, 0)} methods</span>
        </div>
      </div>
    </div>
  );
};
