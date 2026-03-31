export interface QueryPlan {
    id: number;
    database_id: number;
    query_hash: string;
    query_text: string;
    plan_json: Record<string, any>;
    mean_time: number;
    total_time: number;
    calls: number;
    created_at: string;
}

export interface QueryIssue {
    id: number;
    query_plan_id: number;
    issue_type: 'sequential_scan' | 'nested_loop' | 'missing_index' | 'hash_aggregate';
    severity: 'low' | 'medium' | 'high' | 'critical';
    affected_node_id: number;
    description: string;
    recommendation: string;
    estimated_benefit: number;
}

export interface PerformanceTimeline {
    id: number;
    query_plan_id: number;
    metric_timestamp: string;
    avg_duration: number;
    max_duration: number;
    executions: number;
}

export interface QueryPerformanceData {
    queries: QueryPlan[];
    issues: QueryIssue[];
    timeline: PerformanceTimeline[];
}
