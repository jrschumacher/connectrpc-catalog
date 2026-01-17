/**
 * Global setup for full-stack Playwright tests
 * Loads proto definitions via the Go backend (port 8080)
 */

async function globalSetup() {
  // Load protos via the catalog API (Go backend at port 8080)
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

  console.log(`[Full-Stack Setup] Loaded ${result.serviceCount} services from ${result.fileCount} files`);

  // Verify the Eliza service is available at localhost:50051
  try {
    const healthCheck = await fetch('http://localhost:50051/health', {
      method: 'GET',
    });
    if (healthCheck.ok) {
      console.log('[Full-Stack Setup] Eliza test server is healthy');
    }
  } catch (error) {
    console.warn('[Full-Stack Setup] Warning: Could not reach Eliza test server at localhost:50051');
  }
}

export default globalSetup;
