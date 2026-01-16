import { createPromiseClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CatalogService } from '@/gen/catalog/v1/catalog_connect';

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
});

export const catalogClient = createPromiseClient(CatalogService, transport);
