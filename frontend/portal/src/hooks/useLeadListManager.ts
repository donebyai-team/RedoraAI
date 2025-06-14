import { useEffect, useRef } from "react";
import { useLeadFetcher, UseLeadFetcherProps, } from "./useLeadFetcher";

export interface UseLeadListManagerParams extends UseLeadFetcherProps {
    pageNo: number;
}

export interface loadMoreLeadsProps {
    isFetchingMore: boolean;
    hasMore: boolean;
}

export const useLeadListManager = ({
    relevancyScore,
    subReddit,
    dateRange,
    leadStatusFilter,
    leadList,
    setPageNo,
    setHasMore,
    setIsFetchingMore,
    pageNo,
}: UseLeadListManagerParams) => {
    const hasRunInitialFetch = useRef(false);

    const { fetchLeads } = useLeadFetcher({
        relevancyScore,
        subReddit,
        dateRange,
        leadStatusFilter,
        leadList,
        setPageNo,
        setHasMore,
        setIsFetchingMore,
    });

    // initial prioritized fetch
    useEffect(() => {
        if (!hasRunInitialFetch.current) {
            fetchLeads({ fetchType: "initial", shouldFallbackToCompletedLeads: true });
            hasRunInitialFetch.current = true;
        } else {
            fetchLeads({ fetchType: "initial", shouldFallbackToCompletedLeads: false });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [leadStatusFilter, relevancyScore, subReddit, dateRange]);

    // function for loading more on scroll
    const loadMoreLeads = async ({ isFetchingMore, hasMore }: loadMoreLeadsProps) => {
        if (isFetchingMore || !hasMore) return;
        await fetchLeads({ pageNo: pageNo + 1, fetchType: "pagination" });
    };

    return {
        loadMoreLeads,
    };
};
