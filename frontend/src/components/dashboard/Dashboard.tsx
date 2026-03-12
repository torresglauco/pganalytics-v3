import React, { useState, useEffect } from 'react'
import { MainLayout } from '../layout/MainLayout'
import { MetricCard } from './MetricCard'
import { ActivityFeed } from './ActivityFeed'
import { DrillDownCard } from './DrillDownCard'
import { CollectorStatusTable } from './CollectorStatusTable'
import { LoadingSpinner } from '../ui/LoadingSpinner'
import { useAuthStore } from '../../stores/authStore'

export const Dashboard: React.FC = () => {
  const { user } = useAuthStore()
  const [isLoading, setIsLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('24h')

  // Mock data - will be replaced with API calls
  const metrics = [
    {
      title: 'Active Collectors',
      value: 12,
      trend: { direction: 'up' as const, percentage: 0, period: '24h' },
      icon: '📡',
    },
    {
      title: 'Critical Alerts',
      value: 3,
      trend: { direction: 'up' as const, percentage: 12, period: '24h' },
      icon: '🚨',
    },
    {
      title: 'Total Errors',
      value: '1,234',
      trend: { direction: 'down' as const, percentage: 5, period: '24h' },
      icon: '❌',
    },
  ]

  const activities = [
    {
      id: '1',
      type: 'error' as const,
      title: 'High error rate detected',
      description: 'prod-db-1: 245 errors in last 5 minutes',
      timestamp: '2 min ago',
    },
    {
      id: '2',
      type: 'warning' as const,
      title: 'Slow query detected',
      description: 'Query took 8.5s on staging-db',
      timestamp: '15 min ago',
    },
    {
      id: '3',
      type: 'success' as const,
      title: 'Backup completed',
      description: 'Daily backup of analytics_prod completed successfully',
      timestamp: '1 hour ago',
    },
  ]

  const collectors = [
    {
      id: '1',
      hostname: 'prod-db-1.aws',
      environment: 'Production',
      status: 'OK' as const,
      last_heartbeat: '2s ago',
      error_count_24h: 23,
    },
    {
      id: '2',
      hostname: 'staging-db.local',
      environment: 'Staging',
      status: 'SLOW' as const,
      last_heartbeat: '1m ago',
      error_count_24h: 5,
    },
    {
      id: '3',
      hostname: 'dev-db-local',
      environment: 'Development',
      status: 'DOWN' as const,
      last_heartbeat: '2h ago',
      error_count_24h: 0,
    },
  ]

  useEffect(() => {
    // Simulate loading
    const timer = setTimeout(() => setIsLoading(false), 500)
    return () => clearTimeout(timer)
  }, [])

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading dashboard..." />
  }

  return (
    <MainLayout>
      <div className="py-6 md:py-8 px-4 md:px-6">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100 mb-1">
            Dashboard
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Welcome back, {user?.name || 'User'}! Here's what's happening with your databases.
          </p>
        </div>

        {/* Time Range Selector */}
        <div className="mb-6 flex gap-2">
          {['24h', '7d', '30d'].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                timeRange === range
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 hover:bg-slate-200 dark:hover:bg-slate-600'
              }`}
            >
              {range}
            </button>
          ))}
        </div>

        {/* Metrics Row */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          {metrics.map((metric, i) => (
            <MetricCard key={i} {...metric} />
          ))}
        </div>

        {/* Activity & Drill-Down Row */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="lg:col-span-1">
            <ActivityFeed activities={activities} />
          </div>

          <div className="lg:col-span-2 space-y-4">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">
              Explore & Analyze
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <DrillDownCard
                icon="📋"
                title="View Logs"
                description="Filter and analyze PostgreSQL logs in detail"
                href="/logs"
              />
              <DrillDownCard
                icon="📈"
                title="View Metrics"
                description="Charts, performance trends, error distribution"
                href="/metrics"
              />
              <DrillDownCard
                icon="🚨"
                title="Manage Alerts"
                description="View active alerts and manage incidents"
                href="/alerts"
              />
              <DrillDownCard
                icon="📊"
                title="Grafana Dashboards"
                description="Custom dashboards from Grafana (embedded)"
                href="/grafana"
              />
            </div>
          </div>
        </div>

        {/* Collector Status Table */}
        <CollectorStatusTable collectors={collectors} />
      </div>
    </MainLayout>
  )
}

Dashboard.displayName = 'Dashboard'
