import { IntegrationState as DootaIntegrationState, IntegrationType } from '@doota/pb/doota/portal/v1/portal_pb';
import { useClientsContext } from '@doota/ui-core/context/ClientContext';
import { useCallback, useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../../../../store/store';
import { setError, setLoading, setSuccess } from '../../../../store/slices/redditIntegrationSlice';

export function useRedditIntegrationStatus() {
    const { portalClient } = useClientsContext();
    const dispatch = useDispatch();
    const { isConnected, loading, error } = useSelector((state: RootState) => state.redditIntegration);
    const hasFetched = useRef(false);

    const fetchStatus = useCallback(async () => {
        dispatch(setLoading());

        try {
            const resp = await portalClient.getIntegration({ type: IntegrationType.REDDIT });
            const result = resp.status === DootaIntegrationState.ACTIVE;
            dispatch(setSuccess(result));
        } catch (err: any) {
            dispatch(setError(err.message || 'Unknown error'));
        } finally {
            hasFetched.current = true;
        }
    }, [dispatch, portalClient]);

    useEffect(() => {
        if (!hasFetched.current) {
            fetchStatus();
        }
    }, [fetchStatus]);

    return {
        isConnected,
        loading,
        error,
        refresh: fetchStatus,
    };
}
