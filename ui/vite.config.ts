import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@gen': path.resolve(__dirname, '../gen'),
    },
  },
  optimizeDeps: {
    include: ['@bufbuild/protobuf'],
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
