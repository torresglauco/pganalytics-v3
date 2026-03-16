# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY frontend /app

# Build application
RUN npm run build

# Production stage
FROM node:18-alpine

WORKDIR /app

# Copy built files from builder
COPY --from=builder /app/dist ./dist

# Create a proxy server configuration using built-in Node modules
RUN cat > /app/proxy.js << 'EOF'
const http = require('http');
const https = require('https');
const path = require('path');
const fs = require('fs');
const url = require('url');

// API configuration
const API_PROTOCOL = process.env.API_PROTOCOL || 'http';
const API_HOST = process.env.API_HOST || 'api';
const API_PORT = process.env.API_PORT || '8080';
const distPath = path.join(__dirname, 'dist');

const server = http.createServer((req, res) => {
  const pathname = url.parse(req.url).pathname;

  // API proxy - forward API requests to backend
  if (pathname.startsWith('/api')) {
    const queryString = req.url.includes('?') ? '?' + req.url.split('?')[1] : '';
    const apiPath = `${pathname}${queryString}`;

    const proxyReq = (API_PROTOCOL === 'https' ? https : http).request({
      hostname: API_HOST,
      port: API_PORT,
      path: apiPath,
      method: req.method,
      headers: {
        ...req.headers,
        'host': `${API_HOST}:${API_PORT}`
      }
    }, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.on('data', (chunk) => {
        if (!res.destroyed) {
          res.write(chunk);
        }
      });
      proxyRes.on('end', () => {
        if (!res.destroyed) {
          res.end();
        }
      });
    });

    proxyReq.on('error', (err) => {
      console.error('Proxy error:', err);
      if (!res.headersSent) {
        res.writeHead(502, { 'Content-Type': 'application/json' });
      }
      if (!res.destroyed) {
        res.end(JSON.stringify({ error: 'Bad Gateway', message: err.message }));
      }
    });

    req.on('data', (chunk) => {
      proxyReq.write(chunk);
    });
    req.on('end', () => {
      proxyReq.end();
    });
    req.on('error', (err) => {
      console.error('Request error:', err);
      proxyReq.destroy();
    });

    return;
  }

  // Serve static files
  let filePath = path.join(distPath, pathname === '/' ? 'index.html' : pathname);

  fs.stat(filePath, (err, stats) => {
    if (err || !stats.isFile()) {
      // Serve index.html for SPA routing
      filePath = path.join(distPath, 'index.html');
    }

    fs.readFile(filePath, (err, content) => {
      if (err) {
        res.writeHead(404, { 'Content-Type': 'text/plain' });
        res.end('404 Not Found');
        return;
      }

      const ext = path.extname(filePath);
      const contentType = {
        '.html': 'text/html',
        '.js': 'application/javascript',
        '.css': 'text/css',
        '.json': 'application/json',
        '.png': 'image/png',
        '.jpg': 'image/jpeg',
        '.gif': 'image/gif',
        '.svg': 'image/svg+xml',
      }[ext] || 'text/plain';

      res.writeHead(200, { 'Content-Type': contentType });
      res.end(content);
    });
  });
});

const PORT = process.env.PORT || 80;
server.listen(PORT, () => {
  console.log(`Frontend server listening on port ${PORT}`);
  console.log(`API Backend: ${API_PROTOCOL}://${API_HOST}:${API_PORT}`);
  console.log(`Static files: ${distPath}`);
});
EOF

# Expose port
EXPOSE 80

# Start the proxy server
CMD ["node", "/app/proxy.js"]
