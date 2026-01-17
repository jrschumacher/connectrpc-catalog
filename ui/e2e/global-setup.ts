/**
 * Global setup for Playwright tests
 * Loads proto definitions before tests run
 */

async function globalSetup() {
  // Load protos via the catalog API
  const response = await fetch('http://localhost:8080/catalog.v1.CatalogService/LoadProtos', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      bufModule: 'buf.build/connectrpc/eliza',
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to load protos: ${response.status} ${response.statusText}`);
  }

  const result = await response.json();
  if (!result.success) {
    throw new Error(`Failed to load protos: ${result.error}`);
  }

  console.log(`Loaded ${result.serviceCount} services from ${result.fileCount} files`);
}

export default globalSetup;
