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

# Install dependencies
RUN npm install -g serve cors-anywhere

# Copy built files from builder
COPY --from=builder /app/dist ./dist

# Create a simple proxy server script
RUN cat > proxy.js << 'EOF'
const http = require('http');
const https = require('https');
const url = require('url');
const path = require('path');
const fs = require('fs');

// API backend
const API_HOST = process.env.VITE_API_BACKEND_HOST || 'backend';
const API_PORT = process.env.VITE_API_BACKEND_PORT || '8080';
const API_PROTOCOL = process.env.VITE_API_BACKEND_PROTOCOL || 'http';

// Static files serving
const distPath = path.join(__dirname, 'dist');

const server = http.createServer((req, res) => {
  // Enable CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  if (req.method === 'OPTIONS') {
    res.writeHead(204);
    res.end();
    return;
  }

  const parsedUrl = url.parse(req.url, true);
  const pathname = parsedUrl.pathname;

  // Proxy API requests
  if (pathname.startsWith('/api/')) {
    const proxyUrl = `${API_PROTOCOL}://${API_HOST}:${API_PORT}${pathname}${parsedUrl.search || ''}`;

    const proxyReq = (API_PROTOCOL === 'https' ? https : http).request(
      proxyUrl,
      {
        method: req.method,
        headers: {
          ...req.headers,
          host: `${API_HOST}:${API_PORT}`,
        },
      },
      (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
      }
    );

    proxyReq.on('error', (err) => {
      console.error('Proxy error:', err);
      res.writeHead(502, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Bad Gateway' }));
    });

    req.pipe(proxyReq);
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

const PORT = process.env.PORT || 3000;
server.listen(PORT, () => {
  console.log(`Frontend server listening on port ${PORT}`);
  console.log(`API Backend: ${API_PROTOCOL}://${API_HOST}:${API_PORT}`);
});
EOF

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -s http://localhost:3000 >/dev/null 2>&1 || exit 1

# Start application with proxy
CMD ["node", "proxy.js"]
