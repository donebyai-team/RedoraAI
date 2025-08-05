import { IntegrationState, IntegrationType } from '@doota/pb/doota/portal/v1/portal_pb';
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
            const integrations = result.integrations || [];
            const isAnyActiveRedditIntegration = integrations.some(
                (integration) =>
                    integration.status === IntegrationState.ACTIVE &&
                    integration.type === IntegrationType.REDDIT
            );

            dispatch(setSuccess(isAnyActiveRedditIntegration));
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