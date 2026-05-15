/**
 * Topology Graph Component
 * Main visualization component for replication topology using @xyflow/react
 */

import React, { useMemo } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
  type NodeTypes,
  type EdgeTypes,
  BackgroundVariant,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type { ReplicationTopology, TopologyNodeData, TopologyEdgeData } from '../../types/replication';
import { TopologyNode } from './TopologyNode';
import { TopologyEdge } from './TopologyEdge';

// CRITICAL: Define nodeTypes and edgeTypes OUTSIDE component to prevent re-renders
const nodeTypes: NodeTypes = {
  topology: TopologyNode,
};

const edgeTypes: EdgeTypes = {
  topology: TopologyEdge,
};

/**
 * Determine node status from state string
 */
const getNodeStatus = (state: string): 'streaming' | 'catchup' | 'down' => {
  if (state === 'streaming') return 'streaming';
  if (state === 'catchup') return 'catchup';
  return 'down';
};

/**
 * Convert ReplicationTopology to React Flow nodes and edges
 * Layout: Primary at top center, standbys below, cascading at bottom
 */
const convertTopologyToFlowElements = (
  topology: ReplicationTopology
): { nodes: Node<TopologyNodeData>[]; edges: Edge<TopologyEdgeData>[] } => {
  const nodes: Node<TopologyNodeData>[] = [];
  const edges: Edge<TopologyEdgeData>[] = [];

  // Track all collector IDs and their positions
  const nodePositions: Map<string, { x: number; y: number; role: string }> = new Map();

  // Add primary/standby node (the queried node)
  const primaryNode: Node<TopologyNodeData> = {
    id: topology.collector_id,
    type: 'topology',
    position: { x: 300, y: 50 },
    data: {
      label: topology.node_role === 'primary' ? 'Primary' : 'Standby',
      role: topology.node_role,
      status: 'streaming',
      lagMs: 0,
      applicationName: topology.node_role === 'primary' ? 'primary' : 'standby',
      clientAddr: '',
      collectorId: topology.collector_id,
    },
  };
  nodes.push(primaryNode);
  nodePositions.set(topology.collector_id, { x: 300, y: 50, role: topology.node_role });

  // Add upstream node if exists (for standbys)
  if (topology.upstream_collector_id && topology.upstream_host) {
    const upstreamId = topology.upstream_collector_id;
    const upstreamNode: Node<TopologyNodeData> = {
      id: upstreamId,
      type: 'topology',
      position: { x: 300, y: -150 },
      data: {
        label: topology.upstream_host,
        role: 'primary',
        status: 'streaming',
        lagMs: 0,
        applicationName: 'upstream',
        clientAddr: '',
        collectorId: upstreamId,
      },
    };
    nodes.push(upstreamNode);
    nodePositions.set(upstreamId, { x: 300, y: -150, role: 'primary' });

    // Edge from upstream to current node
    edges.push({
      id: `${upstreamId}-${topology.collector_id}`,
      source: upstreamId,
      target: topology.collector_id,
      type: 'topology',
      data: {
        lagMs: 0,
        syncState: 'async',
        state: 'streaming',
      },
    });
  }

  // Add downstream nodes
  if (topology.downstream_nodes && topology.downstream_nodes.length > 0) {
    const startY = 200;
    const spacing = 220;
    const totalWidth = (topology.downstream_nodes.length - 1) * spacing;
    const startX = 300 - totalWidth / 2;

    topology.downstream_nodes.forEach((downstream, index) => {
      const nodeId = downstream.collector_id;
      const x = startX + index * spacing;
      const y = startY;

      const downstreamNode: Node<TopologyNodeData> = {
        id: nodeId,
        type: 'topology',
        position: { x, y },
        data: {
          label: downstream.application_name || `Standby ${index + 1}`,
          role: 'standby',
          status: getNodeStatus(downstream.state),
          lagMs: downstream.replay_lag_ms,
          applicationName: downstream.application_name,
          clientAddr: downstream.client_addr,
          collectorId: nodeId,
        },
      };
      nodes.push(downstreamNode);
      nodePositions.set(nodeId, { x, y, role: 'standby' });

      // Edge from current node to downstream
      edges.push({
        id: `${topology.collector_id}-${nodeId}`,
        source: topology.collector_id,
        target: nodeId,
        type: 'topology',
        data: {
          lagMs: downstream.replay_lag_ms,
          syncState: downstream.sync_state,
          state: downstream.state,
        },
      });
    });
  }

  return { nodes, edges };
};

interface TopologyGraphProps {
  topology: ReplicationTopology;
  className?: string;
}

/**
 * Main topology graph visualization component
 */
export const TopologyGraph: React.FC<TopologyGraphProps> = ({
  topology,
  className,
}) => {
  // Convert topology to nodes and edges with memoization
  const { nodes, edges } = useMemo(
    () => convertTopologyToFlowElements(topology),
    [topology]
  );

  return (
    <div className={className} style={{ width: '100%', height: '100%', minHeight: '400px' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'topology',
        }}
      >
        <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
        <Controls showInteractive={false} />
        <MiniMap
          nodeColor={(node) => {
            const data = node.data as TopologyNodeData;
            if (data?.role === 'primary') return '#10b981';
            if (data?.role === 'standby') return '#3b82f6';
            return '#f59e0b';
          }}
          maskColor="rgba(0, 0, 0, 0.1)"
          style={{ background: 'rgba(255, 255, 255, 0.8)' }}
        />
      </ReactFlow>
    </div>
  );
};

TopologyGraph.displayName = 'TopologyGraph';

export default TopologyGraph;