import { useRef, useCallback } from "react";
import { toast } from "@/components/ui/use-toast";
import { Lead, LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useAppDispatch } from "@/store/hooks";
import { portalClient } from "@/services/grpc";
import { setError, setIsLoading, setLeadList, setLeadStatusFilter } from "@/store/Lead/leadSlice";
import { DEFAULT_DATA_LIMIT } from "@/utils/constants";
import { DateRangeFilter } from "@doota/pb/doota/portal/v1/portal_pb";

export interface FetchFilters {
    status: LeadStatus | null;
    relevancyScore?: number;
    subReddit?: string;
    dateRange?: DateRangeFilter;
    pageCount?: number;
    pageNo?: number;
}

export interface FetchLeadsOptions {
    pageNo?: number;
    useStatusPriority?: boolean;
    fetchType?: "initial" | "pagination";
}

export interface UseLeadFetcherProps {
    relevancyScore?: number;
    subReddit?: string;
    dateRange?: DateRangeFilter;
    leadStatusFilter: LeadStatus | null;
    leadList: Lead[];
    setPageNo: (n: number) => void;
    setCounts?: (data: any) => void;
    setHasMore: (b: boolean) => void;
    setIsFetchingMore: (b: boolean) => void;
}

export const useLeadFetcher = ({
    relevancyScore,
    subReddit,
    dateRange,
    leadStatusFilter,
    leadList,
    setPageNo,
    setCounts,
    setHasMore,
    setIsFetchingMore,
}: UseLeadFetcherProps) => {
    const dispatch = useAppDispatch();
    const hasFetchedPrioritizedLeads = useRef(false);

    // GRPC call wrapper
    const fetchLeadsFromAPI = useCallback(
        async ({
            status,
            relevancyScore,
            subReddit,
            dateRange,
            pageCount = DEFAULT_DATA_LIMIT,
            pageNo = 1,
        }: FetchFilters) => {
            return await portalClient.getRelevantLeads({
                ...(relevancyScore && { relevancyScore }),
                ...(subReddit && { subReddit }),
                ...(status && { status }),
                ...(dateRange && { dateRange }),
                pageCount,
                pageNo,
            });
        },
        []
    );

    // Main fetch function with optional prioritization & pagination
    const fetchLeads = useCallback(
        async ({
            pageNo = 1,
            useStatusPriority = false,
            fetchType = "initial",
        }: FetchLeadsOptions) => {
            try {
                if (fetchType === "initial") {
                    dispatch(setIsLoading(true));
                } else {
                    setIsFetchingMore(true);
                }

                // Prioritized status fetch (runs only once)
                if (useStatusPriority && !hasFetchedPrioritizedLeads.current) {
                    const prioritizedStatuses: LeadStatus[] = [
                        LeadStatus.NEW,
                        LeadStatus.COMPLETED,
                    ];

                    for (const status of prioritizedStatuses) {
                        try {
                            const result = await fetchLeadsFromAPI({
                                status,
                                relevancyScore,
                                subReddit,
                                dateRange,
                                pageNo,
                            });

                            const leads = result?.leads ?? [];
                            const hasMoreResults = leads.length === DEFAULT_DATA_LIMIT;

                            if (leads.length > 0) {
                                dispatch(setLeadList(leads));
                                dispatch(setLeadStatusFilter(status));
                                setCounts?.(result.analysis);
                                setHasMore(hasMoreResults);
                                break;
                            }
                        } catch (err: any) {
                            const message = err?.response?.data?.message || err.message || "Something went wrong";
                            toast({ title: "Error", description: message });
                            dispatch(setError(message));
                        }
                    }

                    hasFetchedPrioritizedLeads.current = true;
                } else {
                    // Standard fetch
                    const result = await fetchLeadsFromAPI({
                        status: leadStatusFilter,
                        relevancyScore,
                        subReddit,
                        dateRange,
                        pageNo,
                    });

                    const leads = result?.leads ?? [];
                    const hasMoreResults = leads.length === DEFAULT_DATA_LIMIT;

                    dispatch(setLeadList([...leadList, ...leads]));
                    setCounts?.(result.analysis);
                    setHasMore(hasMoreResults);
                    setPageNo(pageNo);
                }
            } catch (err: any) {
                const message = err?.response?.data?.message || err.message || "Something went wrong";
                toast({ title: "Error", description: message });
                dispatch(setError(message));
            } finally {
                if (fetchType === "initial") {
                    dispatch(setIsLoading(false));
                } else {
                    setIsFetchingMore(false);
                }
            }
        },
        [
            dispatch,
            fetchLeadsFromAPI,
            relevancyScore,
            subReddit,
            dateRange,
            leadStatusFilter,
            leadList,
            setCounts,
            setHasMore,
            setPageNo,
            setIsFetchingMore,
        ]
    );

    return { fetchLeads };
};
