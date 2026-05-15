# Phase 13: Frontend UI - Research

**Researched:** 2026-05-15
**Domain:** React Frontend, Graph Visualization, Dashboard UI
**Confidence:** HIGH

## Summary

Phase 13 requires implementing three major UI features: replication topology graph visualization, data classification reports, and host inventory dashboards. The project uses a modern React 18 stack with Vite, Tailwind CSS, Zustand for state management, and Recharts for data visualization. For the replication topology graph (REP-06/UI-01), the recommended solution is @xyflow/react (React Flow v12), which is the current actively maintained version of the popular React Flow library for building interactive node-based graphs.

**Primary recommendation:** Use @xyflow/react 12.x for topology visualization, leverage existing patterns from AlertsDashboard and existing API client structure, and follow the established component architecture with Tailwind CSS styling and Vitest testing.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| React | 18.2.0 | UI framework | Project standard, existing codebase |
| React Router DOM | 6.22.0 | Client-side routing | Existing routing patterns |
| Vite | 5.0.8 | Build tooling | Fast development, existing setup |
| TypeScript | 5.3.3 | Type safety | Existing codebase requirement |
| Tailwind CSS | 3.4.1 | Styling | Existing design system |

### Graph Visualization (NEW)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| @xyflow/react | 12.10.2 | Interactive graph visualization | Industry standard for React node graphs, active maintenance |
| @xyflow/background | 12.x | Grid/dots background | Official addon for topology backgrounds |
| @xyflow/controls | 12.x | Zoom/pan controls | Official addon for navigation |
| @xyflow/minimap | 12.x | Overview minimap | Official addon for large graphs |

### Data Visualization (Existing)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| recharts | 3.7.0 | Charts and gauges | Time-series, bar charts, pie charts |
| framer-motion | 12.34.5 | Animations | Smooth transitions, reveals |

### State & Data
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| zustand | 5.0.11 | Global state | Cross-component state |
| axios | 1.6.2 | HTTP client | API requests (alternatives: fetch) |
| react-hook-form | 7.50.0 | Form handling | Complex forms |
| zod | 3.22.4 | Schema validation | API response validation |
| date-fns | 4.1.0 | Date formatting | Timestamp display |
| lucide-react | 0.308.0 | Icons | UI icons |

### Testing
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| vitest | 1.0.0 | Unit tests | Component tests |
| @testing-library/react | 14.1.2 | Testing utilities | React component testing |
| @playwright/test | 1.59.1 | E2E tests | Full flow testing |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| @xyflow/react | reactflow (v11) | v11 is legacy; @xyflow/react is current actively maintained version |
| @xyflow/react | vis-network | vis-network is lower-level, requires more custom code |
| @xyflow/react | d3-force | d3 is more flexible but requires significantly more development effort |
| recharts | chart.js | recharts is already integrated and React-native |

**Installation:**
```bash
cd frontend
npm install @xyflow/react @xyflow/background @xyflow/controls @xyflow/minimap
```

**Version verification:**
```
@xyflow/react: 12.10.2 (verified 2026-05-15)
reactflow: 11.11.4 (legacy - do not use for new projects)
```

## Architecture Patterns

### Recommended Project Structure
```
frontend/src/
├── pages/
│   ├── ReplicationTopologyPage.tsx    # REP-06, UI-01
│   ├── DataClassificationPage.tsx     # UI-03
│   └── HostInventoryPage.tsx          # UI-04
├── components/
│   ├── topology/                       # Graph visualization
│   │   ├── TopologyGraph.tsx
│   │   ├── TopologyNode.tsx           # Custom node for PostgreSQL instances
│   │   ├── TopologyEdge.tsx           # Custom edge for replication links
│   │   └── TopologyLegend.tsx
│   ├── classification/                 # Data classification
│   │   ├── ClassificationTable.tsx
│   │   ├── ClassificationFilters.tsx
│   │   ├── ClassificationSummary.tsx
│   │   └── PatternBreakdownChart.tsx
│   └── host/                           # Host inventory
│       ├── HostStatusTable.tsx
│       ├── HostMetricsCard.tsx
│       ├── HostDetailPanel.tsx
│       └── HostInventorySummary.tsx
├── api/
│   ├── replicationApi.ts              # Replication endpoints
│   ├── hostApi.ts                     # Host monitoring endpoints
│   └── classificationApi.ts           # Data classification endpoints
├── types/
│   ├── replication.ts
│   ├── host.ts
│   └── classification.ts
└── hooks/
    ├── useReplicationTopology.ts
    ├── useHostMetrics.ts
    └── useClassification.ts
```

### Pattern 1: API Client Pattern
**What:** Consistent API client structure with fetch, credentials, and error handling
**When to use:** All new API endpoints must follow this pattern
**Example:**
```typescript
// Source: frontend/src/api/alertDashboardApi.ts
const API_BASE = '/api/v1';

async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `API error: ${response.status}`);
  }

  return response.json();
}
```

### Pattern 2: Page Component with Loading State
**What:** Standard page structure with loading, error, and data states
**When to use:** All new page components
**Example:**
```typescript
// Source: frontend/src/pages/AlertsDashboard.tsx pattern
export const AlertsDashboard: React.FC = () => {
  const [data, setData] = useState<DataType[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [/* dependencies */]);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const result = await fetchData();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load');
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading..." />;
  }

  // Render content
};
```

### Pattern 3: Custom React Flow Nodes
**What:** Custom node components for topology visualization
**When to use:** Replication topology graph nodes
**Example:**
```typescript
// Custom node for PostgreSQL instance in topology
import { Handle, Position, type NodeProps } from '@xyflow/react';

interface PostgresNodeData {
  label: string;
  role: 'primary' | 'standby' | 'cascading_standby';
  status: 'streaming' | 'catchup' | 'down';
  lagMs: number;
}

export const PostgresNode = ({ data }: NodeProps<PostgresNodeData>) => {
  const borderColor = data.role === 'primary' ? 'border-emerald-500' : 'border-blue-500';
  const statusColor = data.status === 'streaming' ? 'bg-emerald-500' : 'bg-amber-500';

  return (
    <div className={`px-4 py-2 rounded-lg border-2 bg-white shadow-md ${borderColor}`}>
      <Handle type="target" position={Position.Left} />
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${statusColor}`} />
        <span className="font-medium">{data.label}</span>
      </div>
      <div className="text-xs text-gray-500">{data.role}</div>
      {data.role !== 'primary' && (
        <div className="text-xs text-gray-400">Lag: {data.lagMs}ms</div>
      )}
      <Handle type="source" position={Position.Right} />
    </div>
  );
};
```

### Pattern 4: Recharts Integration
**What:** Use existing Recharts for time-series and classification breakdown
**When to use:** Metrics charts, classification pie charts
**Example:**
```typescript
// Source: frontend/src/components/charts/GaugeChart.tsx
import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';

// Classification breakdown pie chart
const patternColors = {
  CPF: '#f59e0b',
  CNPJ: '#10b981',
  EMAIL: '#3b82f6',
  PHONE: '#8b5cf6',
  CREDIT_CARD: '#ef4444',
};
```

### Anti-Patterns to Avoid
- **Mixing reactflow and @xyflow/react:** Use only @xyflow/react - do not mix versions
- **Inline node types:** Define nodeTypes outside component or useMemo to prevent re-renders
- **Uncontrolled graph updates:** Always use controlled nodes/edges state with proper memoization
- **Skipping loading states:** Always show loading indicators during API calls
- **Ignoring CSRF tokens:** POST/PUT/DELETE must include X-CSRF-Token header (see authApi.ts)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Graph layout algorithm | Custom force-directed layout | @xyflow/react with dagre | Edge cases, performance, maintainability |
| Node/edge interactions | Custom drag/zoom logic | @xyflow/react built-in | Complex event handling, accessibility |
| Form validation | Manual validation | react-hook-form + zod | Existing pattern, error handling |
| Table pagination/sorting | Custom table logic | @tanstack/react-table | Already integrated, feature-complete |
| Date formatting | Manual date manipulation | date-fns | Existing dependency, locale support |
| State persistence | Custom localStorage sync | zustand persist middleware | Simpler, tested solution |

**Key insight:** The project already has well-established patterns for API calls, forms, tables, and state management. Reusing these ensures consistency and reduces development time.

## Common Pitfalls

### Pitfall 1: @xyflow/react Node Type Re-renders
**What goes wrong:** Defining nodeTypes inside component causes constant re-renders and graph flickering
**Why it happens:** React Flow compares nodeTypes by reference, not value
**How to avoid:** Define nodeTypes outside component or use useMemo
**Warning signs:** Graph flickers on every parent state change, poor performance

```typescript
// WRONG
const MyComponent = () => {
  const nodeTypes = { postgres: PostgresNode }; // Recreated every render
  return <ReactFlow nodeTypes={nodeTypes} />;
};

// CORRECT
const nodeTypes = { postgres: PostgresNode }; // Outside component

const MyComponent = () => {
  return <ReactFlow nodeTypes={nodeTypes} />;
};
```

### Pitfall 2: Missing CSRF Token
**What goes wrong:** POST/PUT/DELETE requests fail with 401/403
**Why it happens:** Backend requires CSRF token for state-changing operations
**How to avoid:** Follow authApi.ts pattern - extract CSRF token from cookie and add to headers
**Warning signs:** 401 errors on form submissions

### Pitfall 3: Uncontrolled Graph State
**What goes wrong:** Nodes/edges don't update when data changes
**Why it happens:** React Flow can work in controlled or uncontrolled mode
**How to avoid:** Use controlled nodes/edges with useState/useMemo and update via setNodes/setEdges
**Warning signs:** Graph shows stale data after API updates

### Pitfall 4: Large Topology Performance
**What goes wrong:** Graph becomes slow with many nodes
**Why it happens:** React Flow renders all nodes, not just visible ones
**How to avoid:** Use `useMemo` for node/edge creation, consider virtualization for 100+ nodes
**Warning signs:** Laggy dragging, slow zoom on large topologies

### Pitfall 5: Classification Drill-Down State
**What goes wrong:** Filter state lost when navigating between database/table views
**Why it happens:** State not persisted in URL or global store
**How to avoid:** Use URL query params for filter state or zustand store
**Warning signs:** Back button loses filter context

## Code Examples

### Replication Topology API Client
```typescript
// Source: Based on existing API patterns
// frontend/src/api/replicationApi.ts
import type { ReplicationTopology, ReplicationStatus, ReplicationSlot } from '../types/replication';

const API_BASE = '/api/v1';

async function apiCall<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `API error: ${response.status}`);
  }

  return response.json();
}

export const replicationApi = {
  getTopology: async (collectorId: string): Promise<ReplicationTopology> => {
    const response = await apiCall<{ topology: ReplicationTopology }>(
      `/collectors/${collectorId}/topology`
    );
    return response.topology;
  },

  getReplicationStatus: async (collectorId: string): Promise<ReplicationStatus[]> => {
    const response = await apiCall<{ data: ReplicationStatus[] }>(
      `/collectors/${collectorId}/replication`
    );
    return response.data;
  },

  getReplicationSlots: async (collectorId: string): Promise<ReplicationSlot[]> => {
    const response = await apiCall<{ data: ReplicationSlot[] }>(
      `/collectors/${collectorId}/replication-slots`
    );
    return response.data;
  },
};
```

### Host Status API Client
```typescript
// frontend/src/api/hostApi.ts
import type { HostStatus, HostMetrics, HostInventory } from '../types/host';

const API_BASE = '/api/v1';

async function apiCall<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `API error: ${response.status}`);
  }

  return response.json();
}

export const hostApi = {
  getAllHostStatuses: async (threshold = 300): Promise<HostStatus[]> => {
    const response = await apiCall<{ status: HostStatus[] }>(
      `/hosts?threshold=${threshold}`
    );
    return response.status;
  },

  getHostMetrics: async (collectorId: string, timeRange = '24h'): Promise<HostMetrics[]> => {
    const response = await apiCall<{ data: HostMetrics[] }>(
      `/hosts/${collectorId}/metrics?time_range=${timeRange}`
    );
    return response.data;
  },

  getHostInventory: async (collectorId: string): Promise<HostInventory> => {
    const response = await apiCall<{ data: HostInventory }>(
      `/hosts/${collectorId}/inventory`
    );
    return response.data;
  },
};
```

### Classification API Client
```typescript
// frontend/src/api/classificationApi.ts
import type { DataClassificationResult, ClassificationReportResponse } from '../types/classification';

const API_BASE = '/api/v1';

export const classificationApi = {
  getClassificationResults: async (
    collectorId: string,
    filters: {
      database?: string;
      schema?: string;
      table?: string;
      patternType?: string;
      category?: string;
      timeRange?: string;
    } = {}
  ): Promise<DataClassificationResult[]> => {
    const params = new URLSearchParams();
    if (filters.database) params.append('database', filters.database);
    if (filters.schema) params.append('schema', filters.schema);
    if (filters.table) params.append('table', filters.table);
    if (filters.patternType) params.append('pattern_type', filters.patternType);
    if (filters.category) params.append('category', filters.category);
    if (filters.timeRange) params.append('time_range', filters.timeRange);

    const response = await apiCall<{ data: DataClassificationResult[] }>(
      `/collectors/${collectorId}/classification?${params.toString()}`
    );
    return response.data;
  },

  getClassificationReport: async (collectorId: string): Promise<ClassificationReportResponse> => {
    return apiCall<ClassificationReportResponse>(
      `/collectors/${collectorId}/classification/report`
    );
  },
};
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| reactflow v11 | @xyflow/react v12 | 2024 | New package scope, same API, active maintenance |
| Class components | Functional components with hooks | 2020+ | Project standard, simpler code |
| Redux | Zustand | 2022+ | Simpler state, less boilerplate |
| CSS modules | Tailwind CSS | 2022+ | Faster development, consistent design |
| Jest | Vitest | 2023+ | Faster tests, native ESM |

**Deprecated/outdated:**
- **reactflow (v11):** Use @xyflow/react (v12) for new projects - same team, active development
- **Class components:** Project uses functional components with hooks

## Open Questions

1. **Should topology support real-time updates via WebSocket?**
   - What we know: Backend has realtime infrastructure (realtimeClient in App.tsx)
   - What's unclear: Whether replication topology changes frequently enough to warrant real-time
   - Recommendation: Start with polling (30s interval), add WebSocket if users request it

2. **Should classification support custom pattern creation in this phase?**
   - What we know: Backend API supports custom patterns (DATA-04)
   - What's unclear: Whether UI-05 (notification channels) is in scope
   - Recommendation: Focus on visualization (UI-03), defer custom pattern UI to follow-up

3. **What is the maximum expected topology size?**
   - What we know: System supports 2000+ clusters (SCALE-01)
   - What's unclear: Typical topology depth and breadth
   - Recommendation: Design for 20-50 nodes per topology with virtualization support for larger

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest 1.0.0 |
| Config file | frontend/vite.config.ts |
| Quick run command | `npm run test` |
| Full suite command | `npm run test:coverage` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REP-06 | User can view replication topology as interactive graph | unit + e2e | `vitest run src/components/topology/` | Wave 0 |
| UI-01 | User can view replication topology graph | unit | `vitest run src/pages/ReplicationTopologyPage.test.tsx` | Wave 0 |
| UI-03 | User can view data classification reports with drill-down | unit | `vitest run src/pages/DataClassificationPage.test.tsx` | Wave 0 |
| UI-04 | User can view host inventory dashboards with status and metrics | unit | `vitest run src/pages/HostInventoryPage.test.tsx` | Wave 0 |

### Sampling Rate
- **Per task commit:** `npm run test` (quick run)
- **Per wave merge:** `npm run test:coverage` (full suite with coverage)
- **Phase gate:** Full suite green, coverage >= 80% before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `frontend/src/components/topology/TopologyGraph.test.tsx` - covers REP-06 topology rendering
- [ ] `frontend/src/components/topology/TopologyNode.test.tsx` - covers custom node rendering
- [ ] `frontend/src/pages/ReplicationTopologyPage.test.tsx` - covers UI-01 page integration
- [ ] `frontend/src/pages/DataClassificationPage.test.tsx` - covers UI-03 classification page
- [ ] `frontend/src/pages/HostInventoryPage.test.tsx` - covers UI-04 host inventory page
- [ ] `frontend/src/api/replicationApi.test.ts` - covers replication API client
- [ ] `frontend/src/api/hostApi.test.ts` - covers host API client
- [ ] `frontend/src/api/classificationApi.test.ts` - covers classification API client
- [ ] `frontend/src/types/replication.ts` - covers replication type definitions
- [ ] `frontend/src/types/host.ts` - covers host type definitions
- [ ] `frontend/src/types/classification.ts` - covers classification type definitions

## Sources

### Primary (HIGH confidence)
- frontend/package.json - existing dependencies and versions
- backend/internal/api/handlers_replication.go - replication API endpoints
- backend/internal/api/handlers_host.go - host monitoring API endpoints
- backend/internal/api/handlers_data_classification.go - classification API endpoints
- backend/pkg/models/replication_models.go - replication data structures
- backend/pkg/models/host_models.go - host data structures
- backend/pkg/models/classification_models.go - classification data structures

### Secondary (MEDIUM confidence)
- npm registry (@xyflow/react 12.10.2) - current package versions
- frontend/src/api/alertDashboardApi.ts - existing API client patterns
- frontend/src/pages/AlertsDashboard.tsx - existing page patterns
- frontend/src/components/charts/GaugeChart.tsx - existing chart patterns
- frontend/vite.config.ts - test configuration

### Tertiary (LOW confidence)
- Web search results for React Flow best practices - general guidance only

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on existing package.json and verified npm versions
- Architecture: HIGH - Based on existing codebase patterns
- Pitfalls: MEDIUM - Based on React Flow documentation and common issues
- API patterns: HIGH - Based on existing handler implementations

**Research date:** 2026-05-15
**Valid until:** 30 days (stable frontend ecosystem)