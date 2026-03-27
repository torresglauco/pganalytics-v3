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

# Copy proxy server
COPY frontend/proxy.js /app/proxy.js

# Expose port
EXPOSE 80

# Start the proxy server
CMD ["node", "/app/proxy.js"]
