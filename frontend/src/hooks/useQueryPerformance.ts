import { useState, useEffect } from 'react';
import { QueryPerformanceData } from '../types/queryPerformance';
import { apiClient } from '../services/api';

export const useQueryPerformance = (databaseId: string) => {
    const [data, setData] = useState<QueryPerformanceData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const json = await apiClient.getQueryPerformance(databaseId);
                setData(json);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        if (databaseId) {
            fetchData();
        }
    }, [databaseId]);

    return { data, loading, error };
};
