import react from '@vitejs/plugin-react';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig, type Plugin } from 'vite';
import packageJson from './package.json' with { type: 'json' };

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const base = '/group-chat-archaeologist/';

function docsDataServer(): Plugin {
  return {
    name: 'docs-data-server',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (!req.url?.startsWith(`${base}data/`)) {
          next();
          return;
        }
        const relative = decodeURIComponent(req.url.slice(`${base}data/`.length).split('?')[0] ?? '');
        const dataRoot = path.resolve(__dirname, '../docs/data');
        const filePath = path.resolve(dataRoot, relative);
        if (!filePath.startsWith(dataRoot)) {
          res.statusCode = 403;
          res.end('Forbidden');
          return;
        }
        if (!fs.existsSync(filePath) || fs.statSync(filePath).isDirectory()) {
          next();
          return;
        }
        const ext = path.extname(filePath);
        res.setHeader(
          'Content-Type',
          ext === '.svg' ? 'image/svg+xml' : ext === '.json' ? 'application/json' : 'text/plain'
        );
        fs.createReadStream(filePath).pipe(res);
      });
    }
  };
}

export default defineConfig({
  base,
  plugins: [react(), docsDataServer()],
  publicDir: 'public',
  define: {
    __APP_VERSION__: JSON.stringify(packageJson.version)
  },
  build: {
    outDir: '../docs',
    emptyOutDir: false,
    assetsDir: 'assets',
    sourcemap: false,
    rollupOptions: {
      output: {
        assetFileNames: 'assets/[name]-[hash][extname]',
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js'
      }
    }
  }
});
