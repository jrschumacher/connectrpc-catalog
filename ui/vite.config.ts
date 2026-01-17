import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@gen': path.resolve(__dirname, './src/gen'),
      // Resolve protobuf packages from ui/node_modules for gen directory
      '@bufbuild/protobuf': path.resolve(__dirname, 'node_modules/@bufbuild/protobuf'),
      '@connectrpc/connect': path.resolve(__dirname, 'node_modules/@connectrpc/connect'),
    },
  },
  optimizeDeps: {
    include: ['@bufbuild/protobuf', '@connectrpc/connect'],
  },
  server: {
    port: 5173,
    proxy: {
      '/catalog.v1.CatalogService': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    commonjsOptions: {
      include: [/@bufbuild\/protobuf/, /node_modules/],
    },
  },
});
