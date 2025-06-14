import { useEffect, useRef } from "react";
import { useLeadFetcher, UseLeadFetcherProps, } from "./useLeadFetcher";
import { useAppSelector } from "@/store/hooks";

export interface UseLeadListManagerParams extends UseLeadFetcherProps { }

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
    setHasMore,
    setIsFetchingMore,
}: UseLeadListManagerParams) => {
    const hasRunInitialFetch = useRef(false);
    const { pageNo } = useAppSelector((state) => state.lead);
 console.log("####_pageNo ", pageNo)
    const { fetchLeads } = useLeadFetcher({
        relevancyScore,
        subReddit,
        dateRange,
        leadStatusFilter,
        leadList,
        setHasMore,
        setIsFetchingMore,
    });

    // initial prioritized fetch
    useEffect(() => {
        if (!hasRunInitialFetch.current) {
            if (!leadList.length) {
                fetchLeads({ fetchType: "initial", shouldFallbackToCompletedLeads: true });
                hasRunInitialFetch.current = true;
            }
        } else {
            if (!leadList.length) {
                fetchLeads({ fetchType: "initial", shouldFallbackToCompletedLeads: false });
            }
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [leadStatusFilter, relevancyScore, subReddit, dateRange]);

    console.log("####_222")

    // function for loading more on scroll
    const loadMoreLeads = async ({ isFetchingMore, hasMore }: loadMoreLeadsProps) => {
        if (isFetchingMore || !hasMore) return;
        await fetchLeads({ pageNo: pageNo + 1, fetchType: "pagination" });
    };

    return {
        loadMoreLeads,
    };
};
