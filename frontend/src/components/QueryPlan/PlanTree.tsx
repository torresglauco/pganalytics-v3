import React from 'react';
import { QueryPlan } from '../../types/queryPerformance';

interface PlanTreeProps {
    plan: QueryPlan;
}

export const PlanTree: React.FC<PlanTreeProps> = ({ plan }) => {
    const renderNode = (node: any, depth: number = 0) => {
        const indent = depth * 20;

        return (
            <div key={`${node['Node Type']}-${depth}`} style={{ marginLeft: `${indent}px` }} className="p-2 border-l border-gray-300">
                <div className="font-mono text-sm text-gray-700">
                    <span className="font-semibold">{node['Node Type']}</span>
                    {node['Actual Loops'] && ` (loops: ${node['Actual Loops']})`}
                    {node['Total Cost'] && ` [Cost: ${node['Total Cost']}]`}
                </div>
                {node['Plans']?.map((child: any) => renderNode(child, depth + 1))}
            </div>
        );
    };

    return (
        <div className="bg-gray-50 p-4 rounded border border-gray-200">
            <h3 className="font-bold mb-3 text-gray-800">Execution Plan</h3>
            <div className="space-y-1">
                {renderNode(plan.plan_json.Plan)}
            </div>
        </div>
    );
};
