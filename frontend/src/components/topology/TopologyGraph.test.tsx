/**
 * Tests for TopologyGraph component
 * Tests the main visualization component for replication topology using @xyflow/react
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import { TopologyGraph } from './TopologyGraph'
import type { ReplicationTopology } from '../../types/replication'

// Mock @xyflow/react to avoid rendering complexity
vi.mock('@xyflow/react', () => ({
  ReactFlow: ({ nodes, edges, children }: { nodes: any[]; edges: any[]; children?: React.ReactNode }) => (
    <div data-testid="react-flow">
      <div data-testid="nodes">
        {nodes.map((node: any) => (
          <div key={node.id} data-testid={`node-${node.id}`} data-role={node.data.role} data-status={node.data.status}>
            <span data-testid={`node-label-${node.id}`}>{node.data.label}</span>
            <span data-testid={`node-lag-${node.id}`}>{node.data.lagMs}</span>
          </div>
        ))}
      </div>
      <div data-testid="edges">
        {edges.map((edge: any) => (
          <div key={edge.id} data-testid={`edge-${edge.id}`} data-source={edge.source} data-target={edge.target}>
            {edge.data?.lagMs !== undefined && (
              <span data-testid={`edge-lag-${edge.id}`}>{edge.data.lagMs}ms</span>
            )}
          </div>
        ))}
      </div>
      {children}
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: ({ nodeColor }: { nodeColor: (node: any) => string }) => {
    // Test that nodeColor function returns correct colors for different roles
    const primaryColor = nodeColor({ data: { role: 'primary' } })
    const standbyColor = nodeColor({ data: { role: 'standby' } })
    const cascadingColor = nodeColor({ data: { role: 'cascading_standby' } })
    return (
      <div data-testid="minimap">
        <span data-testid="minimap-primary-color">{primaryColor}</span>
        <span data-testid="minimap-standby-color">{standbyColor}</span>
        <span data-testid="minimap-cascading-color">{cascadingColor}</span>
      </div>
    )
  },
  BackgroundVariant: { Dots: 'dots' },
}))

// Mock the custom node and edge components
vi.mock('./TopologyNode', () => ({
  TopologyNode: () => <div data-testid="topology-node" />,
}))

vi.mock('./TopologyEdge', () => ({
  TopologyEdge: () => <div data-testid="topology-edge" />,
}))

describe('TopologyGraph', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock topology data factory
  const createMockTopology = (overrides: Partial<ReplicationTopology> = {}): ReplicationTopology => ({
    collector_id: 'primary-001',
    node_role: 'primary',
    downstream_count: 0,
    ...overrides,
  })

  describe('Primary node rendering', () => {
    it('should render primary node with correct label', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      expect(screen.getByTestId('node-primary-001')).toBeInTheDocument()
      expect(screen.getByTestId('node-label-primary-001')).toHaveTextContent('Primary')
    })

    it('should render primary node with primary role', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      const node = screen.getByTestId('node-primary-001')
      expect(node).toHaveAttribute('data-role', 'primary')
    })

    it('should render primary node with streaming status', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      const node = screen.getByTestId('node-primary-001')
      expect(node).toHaveAttribute('data-status', 'streaming')
    })
  })

  describe('Standby nodes rendering', () => {
    it('should render standby nodes below primary', () => {
      const topology = createMockTopology({
        collector_id: 'primary-001',
        node_role: 'primary',
        downstream_count: 2,
        downstream_nodes: [
          {
            collector_id: 'standby-001',
            application_name: 'standby1',
            client_addr: '192.168.1.10',
            state: 'streaming',
            sync_state: 'async',
            replay_lag_ms: 150,
          },
          {
            collector_id: 'standby-002',
            application_name: 'standby2',
            client_addr: '192.168.1.11',
            state: 'streaming',
            sync_state: 'sync',
            replay_lag_ms: 50,
          },
        ],
      })
      render(<TopologyGraph topology={topology} />)

      expect(screen.getByTestId('node-standby-001')).toBeInTheDocument()
      expect(screen.getByTestId('node-standby-002')).toBeInTheDocument()
    })

    it('should display correct labels for standby nodes', () => {
      const topology = createMockTopology({
        collector_id: 'primary-001',
        node_role: 'primary',
        downstream_count: 1,
        downstream_nodes: [
          {
            collector_id: 'standby-001',
            application_name: 'MyStandby',
            client_addr: '192.168.1.10',
            state: 'streaming',
            sync_state: 'async',
            replay_lag_ms: 150,
          },
        ],
      })
      render(<TopologyGraph topology={topology} />)

      expect(screen.getByTestId('node-label-standby-001')).toHaveTextContent('MyStandby')
    })
  })

  describe('Edges rendering', () => {
    it('should render edges connecting nodes', () => {
      const topology = createMockTopology({
        collector_id: 'primary-001',
        node_role: 'primary',
        downstream_count: 2,
        downstream_nodes: [
          {
            collector_id: 'standby-001',
            application_name: 'standby1',
            client_addr: '192.168.1.10',
            state: 'streaming',
            sync_state: 'async',
            replay_lag_ms: 150,
          },
          {
            collector_id: 'standby-002',
            application_name: 'standby2',
            client_addr: '192.168.1.11',
            state: 'streaming',
            sync_state: 'sync',
            replay_lag_ms: 50,
          },
        ],
      })
      render(<TopologyGraph topology={topology} />)

      const edge1 = screen.getByTestId('edge-primary-001-standby-001')
      const edge2 = screen.getByTestId('edge-primary-001-standby-002')

      expect(edge1).toBeInTheDocument()
      expect(edge2).toBeInTheDocument()
      expect(edge1).toHaveAttribute('data-source', 'primary-001')
      expect(edge1).toHaveAttribute('data-target', 'standby-001')
    })

    it('should display lag metrics on edges', () => {
      const topology = createMockTopology({
        collector_id: 'primary-001',
        node_role: 'primary',
        downstream_count: 1,
        downstream_nodes: [
          {
            collector_id: 'standby-001',
            application_name: 'standby1',
            client_addr: '192.168.1.10',
            state: 'streaming',
            sync_state: 'async',
            replay_lag_ms: 150,
          },
        ],
      })
      render(<TopologyGraph topology={topology} />)

      expect(screen.getByTestId('edge-lag-primary-001-standby-001')).toHaveTextContent('150ms')
    })
  })

  describe('Node colors', () => {
    it('should show correct node colors for primary (green)', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      // Primary should be green (#10b981)
      expect(screen.getByTestId('minimap-primary-color')).toHaveTextContent('#10b981')
    })

    it('should show correct node colors for standby (blue)', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      // Standby should be blue (#3b82f6)
      expect(screen.getByTestId('minimap-standby-color')).toHaveTextContent('#3b82f6')
    })
  })

  describe('Empty topology handling', () => {
    it('should handle empty topology gracefully', () => {
      const topology = createMockTopology({
        collector_id: 'primary-001',
        node_role: 'primary',
        downstream_count: 0,
        downstream_nodes: [],
      })
      render(<TopologyGraph topology={topology} />)

      // Should render the graph container
      expect(screen.getByTestId('react-flow')).toBeInTheDocument()
      // Should have only the primary node
      expect(screen.getByTestId('node-primary-001')).toBeInTheDocument()
      // Should have no edges
      const edges = screen.queryAllByTestId(/edge-/)
      expect(edges).toHaveLength(0)
    })
  })

  describe('MiniMap rendering', () => {
    it('should render MiniMap with correct node colors', () => {
      const topology = createMockTopology({ node_role: 'primary' })
      render(<TopologyGraph topology={topology} />)

      expect(screen.getByTestId('minimap')).toBeInTheDocument()
      expect(screen.getByTestId('minimap-primary-color')).toHaveTextContent('#10b981')
      expect(screen.getByTestId('minimap-standby-color')).toHaveTextContent('#3b82f6')
    })
  })
})