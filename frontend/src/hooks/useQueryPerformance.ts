import { useState, useEffect } from 'react';
import { QueryPerformanceData } from '../types/queryPerformance';

export const useQueryPerformance = (databaseId: string) => {
    const [data, setData] = useState<QueryPerformanceData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch(`/api/v1/query-performance/database/${databaseId}`);
                if (!response.ok) throw new Error('Failed to fetch');
                const json = await response.json();
                setData(json);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [databaseId]);

    return { data, loading, error };
};
