import React from 'react';
import { useParams } from 'react-router-dom';
import { useQueryPerformance } from '../hooks/useQueryPerformance';
import { PlanTree } from '../components/QueryPlan/PlanTree';
import { Timeline } from '../components/QueryPerformance/Timeline';
import { MainLayout } from '../components/layout/MainLayout';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';

export const QueryPerformancePage: React.FC = () => {
    const { databaseId } = useParams<{ databaseId: string }>();
    const { data, loading, error } = useQueryPerformance(databaseId || '');

    if (loading) {
        return (
            <MainLayout>
                <LoadingSpinner fullScreen message="Loading query performance data..." />
            </MainLayout>
        );
    }

    if (error) {
        return (
            <MainLayout>
                <div className="space-y-6">
                    <div>
                        <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
                            Query Performance
                        </h1>
                        <p className="mt-2 text-slate-600 dark:text-slate-400">
                            Analyze and optimize slow queries
                        </p>
                    </div>
                    <div className="p-4 bg-red-50 border border-red-200 rounded text-red-700">
                        Error: {error}
                    </div>
                </div>
            </MainLayout>
        );
    }

    if (!data || !data.queries || data.queries.length === 0) {
        return (
            <MainLayout>
                <div className="space-y-6">
                    <div>
                        <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
                            Query Performance
                        </h1>
                        <p className="mt-2 text-slate-600 dark:text-slate-400">
                            Analyze and optimize slow queries
                        </p>
                    </div>
                    <div className="p-4 bg-blue-50 border border-blue-200 rounded text-blue-700">
                        No query data available
                    </div>
                </div>
            </MainLayout>
        );
    }

    return (
        <MainLayout>
            <div className="space-y-6">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
                        Query Performance
                    </h1>
                    <p className="mt-2 text-slate-600 dark:text-slate-400">
                        Analyze and optimize slow queries
                    </p>
                </div>

                <div className="space-y-6">
                    {data.queries.map(query => (
                        <div key={query.id} className="bg-white dark:bg-slate-800 p-6 rounded-lg shadow border border-slate-200 dark:border-slate-700">
                            <div className="mb-4">
                                <h2 className="text-xl font-bold text-slate-900 dark:text-white">
                                    {query.query_text.substring(0, 100)}
                                    {query.query_text.length > 100 ? '...' : ''}
                                </h2>
                                <div className="text-sm text-slate-600 dark:text-slate-400 mt-2 grid grid-cols-3 gap-4">
                                    <div>
                                        <span className="font-semibold">Calls:</span> {query.calls}
                                    </div>
                                    <div>
                                        <span className="font-semibold">Avg Time:</span> {query.mean_time.toFixed(2)}ms
                                    </div>
                                    <div>
                                        <span className="font-semibold">Total Time:</span> {query.total_time.toFixed(2)}ms
                                    </div>
                                </div>
                            </div>

                            {/* Execution Plan */}
                            <div className="mb-6">
                                <PlanTree plan={query} />
                            </div>

                            {/* Timeline Chart */}
                            {data.timeline && data.timeline.filter(t => t.query_plan_id === query.id).length > 0 && (
                                <div className="mt-6">
                                    <h3 className="font-bold text-lg mb-3 text-slate-800 dark:text-slate-200">Performance Timeline</h3>
                                    <Timeline data={data.timeline.filter(t => t.query_plan_id === query.id)} />
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>
        </MainLayout>
    );
};
