import { renderHook, waitFor } from '@testing-library/react';
import { useQueryPerformance } from './useQueryPerformance';

describe('useQueryPerformance', () => {
    it('fetches query performance data', async () => {
        const { result } = renderHook(() => useQueryPerformance('1'));

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.data).toBeDefined();
    });
});
