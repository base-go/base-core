import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, 'src/main.js'),
      name: 'BaseUI',
      fileName: 'baseui',
      formats: ['es', 'umd']
    },
    outDir: '../static/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          if (assetInfo.name === 'style.css') {
            return 'baseui.css';
          }
          return assetInfo.name;
        }
      }
    },
    sourcemap: true
  },
  server: {
    port: 5173,
    cors: true,
    hmr: {
      port: 5173
    }
  },
  css: {
    postcss: './postcss.config.js'
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development')
  }
});