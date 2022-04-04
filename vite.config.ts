/**
 * @type {import('vite').UserConfig}
 */
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'cmd/web/dist',
    sourcemap: true,
    manifest: true,
    rollupOptions: {
      input: {
        main: 'src/main.ts',
      },
    },
  },
  server: {
    proxy: {
      '/process-login': {
        target: 'http://localhost:4000/',
      },
      '/logout': {
        target: 'http://localhost:3000/',
        rewrite: (path) => '/',
      },
      '/api': {
        target: 'http://localhost:4001',
        changeOrigin: true,
      },
    },
  },
});
