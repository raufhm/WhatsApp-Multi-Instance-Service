import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

// https://vite.js.org/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  base: '/dashboard/',
  build: {
    outDir: '../backend/dist',
    emptyOutDir: true,
    assetsDir: 'assets',
    rollupOptions: {
      output: {
        entryFileNames: `assets/[name].[hash].js`,
        chunkFileNames: `assets/[name].[hash].js`,
        assetFileNames: `assets/[name].[hash].[ext]`
      }
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/dashboard/api': 'http://localhost:8080/dashboard/api'
    }
  }
});