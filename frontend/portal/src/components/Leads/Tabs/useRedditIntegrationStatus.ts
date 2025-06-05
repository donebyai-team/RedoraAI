import { IntegrationState as DootaIntegrationState } from '@doota/pb/doota/portal/v1/portal_pb';
import { useClientsContext } from '@doota/ui-core/context/ClientContext';
import { useCallback, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../../../../store/store';
import { setError, setLoading, setSuccess } from '../../../../store/slices/redditIntegrationSlice';

export function useRedditIntegrationStatus() {
    const { portalClient } = useClientsContext();
    const dispatch = useDispatch();
    const { isConnected, loading, error } = useSelector((state: RootState) => state.redditIntegration);

    const fetchStatus = useCallback(async () => {
        dispatch(setLoading());

        try {
            const result = await portalClient.getIntegrations({});
            const resp = result.integrations?.[0];
            console.log("integration response", resp)
            const status = resp.status === DootaIntegrationState.ACTIVE;
            dispatch(setSuccess(status));
        } catch (err: any) {
            dispatch(setError(err.message || 'Unknown error'));
        }
    }, [dispatch, portalClient]);

    useEffect(() => {
        fetchStatus();
    }, [fetchStatus]);

    return {
        isConnected,
        loading,
        error,
        refresh: fetchStatus,
    };
}