# pgAnalytics Frontend Phase 2: Core Features Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement three core feature pages (Logs Viewer, Metrics & Analytics, Alert Rules) with API integration, filtering, and professional UI consistent with Phase 1 foundation.

**Architecture:**
- Logs Viewer: Table with pagination, filtering (level, time range, search), drill-down to details modal
- Metrics & Analytics: Recharts visualizations (line, bar, pie charts) with time range selector and data aggregation
- Alert Rules: Management interface for CRUD operations on alert rules with condition builder
- Notification Channels: Configuration UI for email, Slack, PagerDuty, Webhook channels
- All pages use MainLayout wrapper, integrate with existing API client, follow Phase 1 design system

**Tech Stack:** React 18, TypeScript, Tailwind CSS, Recharts, React Router v6, Axios, React Query (optional, can use Zustand + manual fetching)

---

## File Structure Overview

### New Pages & Components
```
src/
├── pages/
│   ├── LogsPage.tsx                 (New - container page)
│   ├── MetricsPage.tsx              (New - container page)
│   ├── AlertsPage.tsx               (New - container page)
│   └── ChannelsPage.tsx             (New - container page)
├── components/
│   ├── logs/                        (New folder)
│   │   ├── LogsViewer.tsx           (Main logs component)
│   │   ├── LogsTable.tsx            (Table component)
│   │   ├── LogFilters.tsx           (Filter panel)
│   │   ├── LogDetailsModal.tsx      (Details modal)
│   │   └── SearchBar.tsx            (Search input)
│   ├── metrics/                     (New folder)
│   │   ├── MetricsViewer.tsx        (Main metrics component)
│   │   ├── ErrorTrendChart.tsx      (Line chart)
│   │   ├── LogDistributionChart.tsx (Pie chart)
│   │   ├── TopErrorCodesChart.tsx   (Bar chart)
│   │   └── MetricsControls.tsx      (Time range selector)
│   ├── alerts/                      (New folder)
│   │   ├── AlertsViewer.tsx         (Main alerts component)
│   │   ├── AlertsTable.tsx          (Rules table)
│   │   ├── AlertForm.tsx            (Create/edit form)
│   │   ├── ConditionBuilder.tsx     (Condition UI)
│   │   └── AlertDetailsModal.tsx    (Details modal)
│   └── channels/                    (New folder)
│       ├── ChannelsViewer.tsx       (Main channels component)
│       ├── ChannelsTable.tsx        (Channels table)
│       ├── ChannelForm.tsx          (Create/edit form)
│       └── ChannelDetailsModal.tsx  (Details modal)
├── hooks/                           (New folder)
│   ├── useLogs.ts                   (API hook for logs)
│   ├── useMetrics.ts                (API hook for metrics)
│   ├── useAlerts.ts                 (API hook for alerts)
│   └── useChannels.ts               (API hook for channels)
└── services/
    └── api.ts                       (Existing - add new endpoints)
```

### Files to Create (19 new files)
- `src/pages/LogsPage.tsx`
- `src/pages/MetricsPage.tsx`
- `src/pages/AlertsPage.tsx`
- `src/pages/ChannelsPage.tsx`
- `src/components/logs/LogsViewer.tsx`
- `src/components/logs/LogsTable.tsx`
- `src/components/logs/LogFilters.tsx`
- `src/components/logs/LogDetailsModal.tsx`
- `src/components/logs/SearchBar.tsx`
- `src/components/metrics/MetricsViewer.tsx`
- `src/components/metrics/ErrorTrendChart.tsx`
- `src/components/metrics/LogDistributionChart.tsx`
- `src/components/metrics/TopErrorCodesChart.tsx`
- `src/components/metrics/MetricsControls.tsx`
- `src/components/alerts/AlertsViewer.tsx`
- `src/components/alerts/AlertsTable.tsx`
- `src/components/alerts/AlertForm.tsx`
- `src/components/alerts/ConditionBuilder.tsx`
- `src/components/alerts/AlertDetailsModal.tsx`
- `src/components/channels/ChannelsViewer.tsx`
- `src/components/channels/ChannelsTable.tsx`
- `src/components/channels/ChannelForm.tsx`
- `src/components/channels/ChannelDetailsModal.tsx`
- `src/hooks/useLogs.ts`
- `src/hooks/useMetrics.ts`
- `src/hooks/useAlerts.ts`
- `src/hooks/useChannels.ts`

### Files to Modify
- `src/App.tsx` (add new routes)
- `src/services/api.ts` (add new API endpoints)

---

## Chunk 1: API Client Updates & Hooks

### Task 1: Update API Client with New Endpoints

**Files:**
- Modify: `src/services/api.ts`

**Description:** Add API endpoints for logs, metrics, alerts, and channels to the API client service.

- [ ] **Step 1: Read current api.ts**

Check existing structure and patterns.

- [ ] **Step 2: Add endpoint methods to API client**

```typescript
// Add to apiClient object in src/services/api.ts

// Logs endpoints
getLogs: async (params: {
  page?: number
  page_size?: number
  level?: string
  search?: string
  instance_id?: string
  from_time?: string
  to_time?: string
}) => {
  const response = await instance.get('/logs', { params })
  return response.data
},

getLogDetails: async (logId: string) => {
  const response = await instance.get(`/logs/${logId}`)
  return response.data
},

// Metrics endpoints
getMetrics: async (params: {
  instance_id?: string
  time_range?: string // '24h', '7d', '30d'
}) => {
  const response = await instance.get('/metrics', { params })
  return response.data
},

getErrorTrend: async (params: {
  instance_id?: string
  hours?: number
}) => {
  const response = await instance.get('/metrics/error-trend', { params })
  return response.data
},

getLogDistribution: async (params: {
  instance_id?: string
  time_range?: string
}) => {
  const response = await instance.get('/metrics/log-distribution', { params })
  return response.data
},

// Alert endpoints
getAlerts: async (params: {
  page?: number
  page_size?: number
  status?: string
}) => {
  const response = await instance.get('/alerts', { params })
  return response.data
},

createAlert: async (data: any) => {
  const response = await instance.post('/alerts', data)
  return response.data
},

updateAlert: async (alertId: string, data: any) => {
  const response = await instance.put(`/alerts/${alertId}`, data)
  return response.data
},

deleteAlert: async (alertId: string) => {
  await instance.delete(`/alerts/${alertId}`)
},

testAlert: async (alertId: string) => {
  const response = await instance.post(`/alerts/${alertId}/test`)
  return response.data
},

// Channel endpoints
getChannels: async () => {
  const response = await instance.get('/channels')
  return response.data
},

createChannel: async (data: any) => {
  const response = await instance.post('/channels', data)
  return response.data
},

updateChannel: async (channelId: string, data: any) => {
  const response = await instance.put(`/channels/${channelId}`, data)
  return response.data
},

deleteChannel: async (channelId: string) => {
  await instance.delete(`/channels/${channelId}`)
},

testChannel: async (channelId: string) => {
  const response = await instance.post(`/channels/${channelId}/test`)
  return response.data
},
```

- [ ] **Step 3: Verify TypeScript compilation**

```bash
cd frontend && npm run type-check
```

Expected: No TypeScript errors.

- [ ] **Step 4: Commit**

```bash
git add src/services/api.ts
git commit -m "feat: add API endpoints for logs, metrics, alerts, and channels"
```

### Task 2: Create Custom Hooks for Data Fetching

**Files:**
- Create: `src/hooks/useLogs.ts`
- Create: `src/hooks/useMetrics.ts`
- Create: `src/hooks/useAlerts.ts`
- Create: `src/hooks/useChannels.ts`

- [ ] **Step 1: Create useLogs hook**

```typescript
// src/hooks/useLogs.ts
import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface LogsParams {
  page?: number
  page_size?: number
  level?: string
  search?: string
  instance_id?: string
  from_time?: string
  to_time?: string
}

export const useLogs = (params: LogsParams = {}) => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchLogs = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getLogs(params)
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
  }, [JSON.stringify(params)])

  const getLogDetails = async (logId: string) => {
    try {
      const result = await apiClient.getLogDetails(logId)
      return result
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch log details')
      return null
    }
  }

  return { data, loading, error, fetchLogs, getLogDetails }
}
```

- [ ] **Step 2: Create useMetrics hook**

```typescript
// src/hooks/useMetrics.ts
import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface MetricsParams {
  instance_id?: string
  time_range?: string
}

export const useMetrics = (params: MetricsParams = {}) => {
  const [metrics, setMetrics] = useState<any>(null)
  const [errorTrend, setErrorTrend] = useState<any>(null)
  const [distribution, setDistribution] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchMetrics = async () => {
    try {
      setLoading(true)
      setError(null)

      const [metricsData, trendData, distData] = await Promise.all([
        apiClient.getMetrics(params),
        apiClient.getErrorTrend({ instance_id: params.instance_id, hours: 24 }),
        apiClient.getLogDistribution(params),
      ])

      setMetrics(metricsData)
      setErrorTrend(trendData)
      setDistribution(distData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch metrics')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMetrics()
  }, [JSON.stringify(params)])

  return { metrics, errorTrend, distribution, loading, error, fetchMetrics }
}
```

- [ ] **Step 3: Create useAlerts hook**

```typescript
// src/hooks/useAlerts.ts
import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface AlertsParams {
  page?: number
  page_size?: number
  status?: string
}

export const useAlerts = (params: AlertsParams = {}) => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchAlerts = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getAlerts(params)
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch alerts')
    } finally {
      setLoading(false)
    }
  }

  const createAlert = async (alertData: any) => {
    try {
      const result = await apiClient.createAlert(alertData)
      await fetchAlerts()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create alert'
      setError(errorMsg)
      throw err
    }
  }

  const updateAlert = async (alertId: string, alertData: any) => {
    try {
      const result = await apiClient.updateAlert(alertId, alertData)
      await fetchAlerts()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update alert'
      setError(errorMsg)
      throw err
    }
  }

  const deleteAlert = async (alertId: string) => {
    try {
      await apiClient.deleteAlert(alertId)
      await fetchAlerts()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete alert'
      setError(errorMsg)
      throw err
    }
  }

  useEffect(() => {
    fetchAlerts()
  }, [JSON.stringify(params)])

  return { data, loading, error, fetchAlerts, createAlert, updateAlert, deleteAlert }
}
```

- [ ] **Step 4: Create useChannels hook**

```typescript
// src/hooks/useChannels.ts
import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

export const useChannels = () => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchChannels = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getChannels()
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch channels')
    } finally {
      setLoading(false)
    }
  }

  const createChannel = async (channelData: any) => {
    try {
      const result = await apiClient.createChannel(channelData)
      await fetchChannels()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create channel'
      setError(errorMsg)
      throw err
    }
  }

  const updateChannel = async (channelId: string, channelData: any) => {
    try {
      const result = await apiClient.updateChannel(channelId, channelData)
      await fetchChannels()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update channel'
      setError(errorMsg)
      throw err
    }
  }

  const deleteChannel = async (channelId: string) => {
    try {
      await apiClient.deleteChannel(channelId)
      await fetchChannels()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete channel'
      setError(errorMsg)
      throw err
    }
  }

  const testChannel = async (channelId: string) => {
    try {
      const result = await apiClient.testChannel(channelId)
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to test channel'
      setError(errorMsg)
      throw err
    }
  }

  useEffect(() => {
    fetchChannels()
  }, [])

  return { data, loading, error, fetchChannels, createChannel, updateChannel, deleteChannel, testChannel }
}
```

- [ ] **Step 5: Verify TypeScript compilation**

```bash
cd frontend && npm run type-check
```

Expected: No TypeScript errors.

- [ ] **Step 6: Commit**

```bash
git add src/hooks/
git commit -m "feat: create custom hooks for logs, metrics, alerts, and channels data fetching"
```

---

## Chunk 2: Logs Viewer Page (5 Components)

### Task 3: Create Logs Components (LogsViewer, LogsTable, LogFilters, LogDetailsModal, SearchBar)

**Files:**
- Create: `src/components/logs/LogsViewer.tsx`
- Create: `src/components/logs/LogsTable.tsx`
- Create: `src/components/logs/LogFilters.tsx`
- Create: `src/components/logs/LogDetailsModal.tsx`
- Create: `src/components/logs/SearchBar.tsx`
- Create: `src/pages/LogsPage.tsx`

This is a large task. Implementation details provided in next message section.

---

## Chunk 3: Metrics & Analytics Page (5 Components)

### Task 4: Create Metrics Components (MetricsViewer, Charts, Controls)

**Files:**
- Create: `src/components/metrics/MetricsViewer.tsx`
- Create: `src/components/metrics/ErrorTrendChart.tsx`
- Create: `src/components/metrics/LogDistributionChart.tsx`
- Create: `src/components/metrics/TopErrorCodesChart.tsx`
- Create: `src/components/metrics/MetricsControls.tsx`
- Create: `src/pages/MetricsPage.tsx`

---

## Chunk 4: Alert Rules Management (5 Components)

### Task 5: Create Alerts Components (AlertsViewer, AlertsTable, AlertForm, ConditionBuilder, AlertDetailsModal)

**Files:**
- Create: `src/components/alerts/AlertsViewer.tsx`
- Create: `src/components/alerts/AlertsTable.tsx`
- Create: `src/components/alerts/AlertForm.tsx`
- Create: `src/components/alerts/ConditionBuilder.tsx`
- Create: `src/components/alerts/AlertDetailsModal.tsx`
- Create: `src/pages/AlertsPage.tsx`

---

## Chunk 5: Notification Channels (4 Components)

### Task 6: Create Channels Components (ChannelsViewer, ChannelsTable, ChannelForm, ChannelDetailsModal)

**Files:**
- Create: `src/components/channels/ChannelsViewer.tsx`
- Create: `src/components/channels/ChannelsTable.tsx`
- Create: `src/components/channels/ChannelForm.tsx`
- Create: `src/components/channels/ChannelDetailsModal.tsx`
- Create: `src/pages/ChannelsPage.tsx`

---

## Chunk 6: Routing Updates

### Task 7: Update App.tsx with New Routes

**Files:**
- Modify: `src/App.tsx`

- [ ] **Step 1: Add new imports**

```typescript
import { LogsPage } from './pages/LogsPage'
import { MetricsPage } from './pages/MetricsPage'
import { AlertsPage } from './pages/AlertsPage'
import { ChannelsPage } from './pages/ChannelsPage'
```

- [ ] **Step 2: Add new routes inside protected routes**

```typescript
{isAuthenticated ? (
  <>
    <Route path="/" element={<Dashboard />} />
    <Route path="/logs" element={<LogsPage />} />
    <Route path="/metrics" element={<MetricsPage />} />
    <Route path="/alerts" element={<AlertsPage />} />
    <Route path="/channels" element={<ChannelsPage />} />
    <Route path="*" element={<Navigate to="/" />} />
  </>
) : (
  <Route path="*" element={<Navigate to="/login" />} />
)}
```

- [ ] **Step 3: Verify TypeScript compilation**

```bash
cd frontend && npm run type-check
```

- [ ] **Step 4: Commit**

```bash
git add src/App.tsx
git commit -m "feat: add routes for logs, metrics, alerts, and channels pages"
```

---

## Implementation Summary

**Phase 2 implements 4 core feature pages:**

1. **Logs Viewer** (5 components)
   - Main table with pagination, sorting, filtering
   - Advanced filters (level, time range, search, instance)
   - Log details modal with full context
   - Professional styling with badges for log levels

2. **Metrics & Analytics** (5 components)
   - Multiple Recharts visualizations
   - Error trend line chart (24h)
   - Log distribution pie chart
   - Top error codes bar chart
   - Time range selector (24h, 7d, 30d, custom)
   - Responsive grid layout

3. **Alert Rules Management** (5 components)
   - CRUD operations for alert rules
   - Condition builder for flexible rule creation
   - Status toggle (enabled/disabled)
   - Test alert functionality
   - Modal for details and editing

4. **Notification Channels** (4 components)
   - Support for Email, Slack, PagerDuty, Webhook, Jira
   - CRUD operations for channels
   - Test channel functionality
   - Type-specific configuration forms
   - Status indicators

**All components:**
- Follow Phase 1 design system (Tailwind CSS, colors, spacing)
- Support dark mode
- Are fully typed with TypeScript
- Include error handling and loading states
- Use MainLayout wrapper for consistency
- Integrate with API client via custom hooks

**Total deliverables:**
- 4 page containers
- 19 reusable components
- 4 custom data fetching hooks
- Updated API client with 15+ endpoints
- Full routing integration

---

**Plan Status:** ✅ Ready for Implementation
**Next Step:** Use superpowers:subagent-driven-development to execute tasks 3-7
