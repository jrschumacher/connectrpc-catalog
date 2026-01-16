import { useState } from 'react';
import { ChevronRight, ChevronDown, Search } from 'lucide-react';
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
  const [expandedServices, setExpandedServices] = useState<Set<string>>(
    new Set()
  );
  const [searchQuery, setSearchQuery] = useState('');

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

  return (
    <div className="flex h-full flex-col">
      <div className="p-4 border-b">
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search services..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8"
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2">
          {filteredServices.length === 0 ? (
            <div className="p-4 text-sm text-muted-foreground text-center">
              No services found
            </div>
          ) : (
            filteredServices.map((service) => {
              const isExpanded = expandedServices.has(service.fullName);
              return (
                <div key={service.fullName} className="mb-1">
                  <button
                    onClick={() => toggleService(service.fullName)}
                    className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm hover:bg-accent"
                  >
                    {isExpanded ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                    <span className="font-medium truncate">{service.name}</span>
                    <span className="ml-auto text-xs text-muted-foreground">
                      {service.methods.length}
                    </span>
                  </button>

                  {isExpanded && (
                    <div className="ml-6 mt-1 space-y-1">
                      {service.methods.map((method) => {
                        const methodId = `${service.fullName}.${method.name}`;
                        const isSelected = selectedMethod === methodId;
                        return (
                          <button
                            key={method.name}
                            onClick={() =>
                              onSelectMethod(service.fullName, method.name)
                            }
                            className={cn(
                              'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm',
                              isSelected
                                ? 'bg-primary text-primary-foreground'
                                : 'hover:bg-accent'
                            )}
                          >
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
    </div>
  );
};
