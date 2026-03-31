import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { PerformanceTimeline } from '../../types/queryPerformance';

interface TimelineProps {
    data: PerformanceTimeline[];
}

export const Timeline: React.FC<TimelineProps> = ({ data }) => {
    const chartData = data.map(d => ({
        timestamp: new Date(d.metric_timestamp).toLocaleTimeString(),
        duration: Math.round(d.avg_duration * 100) / 100,
        executions: d.executions,
    }));

    return (
        <div className="w-full">
            <ResponsiveContainer width="100%" height={300}>
                <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                        dataKey="timestamp"
                        angle={-45}
                        textAnchor="end"
                        height={80}
                    />
                    <YAxis />
                    <Tooltip
                        formatter={(value) => {
                            if (typeof value === 'number') {
                                return [Math.round(value * 100) / 100, 'ms'];
                            }
                            return value;
                        }}
                    />
                    <Legend />
                    <Line
                        type="monotone"
                        dataKey="duration"
                        stroke="#8884d8"
                        name="Avg Duration (ms)"
                        dot={false}
                    />
                </LineChart>
            </ResponsiveContainer>
        </div>
    );
};
