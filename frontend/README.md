# pgAnalytics Collector Management UI

Modern React-based web interface for managing pgAnalytics database collectors.

## Features

- **Register Collectors** - Easy form-based registration with connection testing
- **Manage Collectors** - List, view details, and delete collectors
- **Real-time Status** - Monitor collector health and metrics collection
- **Secure** - JWT token-based authentication with registration secret validation
- **Responsive** - Works on desktop and mobile devices

## Installation

```bash
npm install
```

## Development

Start the development server:

```bash
npm run dev
```

The app will be available at `http://localhost:3000`

Backend API should be running at `http://localhost:8080`

## Building

Build for production:

```bash
npm run build
```

Preview production build:

```bash
npm run preview
```

## Configuration

Create a `.env` file from `.env.example`:

```bash
cp .env.example .env
```

Edit `.env` to set your backend API URL:

```
VITE_API_URL=http://your-backend:8080/api/v1
```

## Usage

### Register a Collector

1. Go to "Register Collector" tab
2. Enter registration secret (from backend config)
3. Enter collector hostname
4. Click "Test" to verify connectivity
5. Fill optional fields (environment, group, description)
6. Click "Register Collector"
7. Copy the generated token and use it in your collector configuration

### Manage Collectors

1. Go to "Manage Collectors" tab
2. View all registered collectors with their status
3. Click refresh to update the list
4. Delete collectors as needed

## Architecture

```
frontend/
├── src/
│   ├── components/          # React components
│   │   ├── CollectorForm.tsx
│   │   └── CollectorList.tsx
│   ├── pages/               # Page components
│   │   └── Dashboard.tsx
│   ├── services/            # API client
│   │   └── api.ts
│   ├── hooks/               # Custom React hooks
│   │   └── useCollectors.ts
│   ├── types/               # TypeScript types
│   │   └── index.ts
│   ├── styles/              # CSS
│   ├── App.tsx              # Root component
│   └── main.tsx             # Entry point
├── public/
│   └── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
└── tailwind.config.js
```

## Technology Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **React Hook Form** - Form handling
- **Zod** - Schema validation
- **Axios** - HTTP client
- **Lucide React** - Icons

## API Integration

The UI communicates with the backend via REST API:

- `POST /api/v1/collectors/register` - Register new collector
- `GET /api/v1/collectors` - List collectors
- `GET /api/v1/collectors/{id}` - Get collector details
- `DELETE /api/v1/collectors/{id}` - Delete collector
- `POST /api/v1/collectors/test-connection` - Test database connection

All requests require Bearer token authentication (from `/api/v1/auth/login`).

## Docker

Build Docker image:

```bash
docker build -f Dockerfile.frontend -t pganalytics-ui:latest .
```

Run with Docker Compose:

```bash
docker-compose up frontend
```

## License

proprietary
