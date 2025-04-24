import { IntegrationState, IntegrationType } from '@doota/pb/doota/portal/v1/portal_pb';
import { useClientsContext } from '@doota/ui-core/context/ClientContext';
import { useEffect, useState } from 'react';

type IntegrationStatus = {
    isConnected: boolean | null;
    loading: boolean;
    error: any;
};

export function useRedditIntegrationStatus(): IntegrationStatus {
    const { portalClient } = useClientsContext();
    const [isConnected, setIsConnected] = useState<boolean | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<any>(null);

    useEffect(() => {
        let isMounted = true;

        portalClient
            .getIntegration({ type: IntegrationType.REDDIT })
            .then((resp) => {
                if (isMounted) {
                    const result = resp.status != IntegrationState.ACTIVE ? false : true;
                    setIsConnected(result);
                    setLoading(false);
                }
            })
            .catch((err) => {
                if (isMounted) {
                    setError(err);
                    setLoading(false);
                }
            });

        return () => {
            isMounted = false;
        };
    }, [portalClient]);

    return { isConnected, loading, error };
}
