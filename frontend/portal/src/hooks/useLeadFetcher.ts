// hooks/useLeadFetcher.ts
import { useRef, useCallback } from "react";
import { toast } from "@/components/ui/use-toast";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useAppDispatch } from "@/store/hooks";
import { portalClient } from "@/services/grpc";
import { setError, setIsLoading, setLeadStatusFilter, setNewTabList } from "@/store/Lead/leadSlice";
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

export interface fetchLeadsProps {
    pageNo?: number;
    append?: boolean;
    usePriority?: boolean;
    loadType?: "initial" | "pagination";
}

export const useLeadFetcher = ({
    relevancyScore,
    subReddit,
    dateRange,
    leadStatusFilter,
    newTabList,
    setPageNo,
    setCounts,
    setHasMore,
    setIsFetchingMore,
}: {
    relevancyScore?: number;
    subReddit?: string;
    dateRange?: any;
    leadStatusFilter: LeadStatus | null;
    newTabList: any[];
    setPageNo: (n: number) => void;
    setCounts?: (data: any) => void;
    setHasMore: (b: boolean) => void;
    setIsFetchingMore: (b: boolean) => void;
}) => {
    const dispatch = useAppDispatch();
    const hasCheckedInitialLeads = useRef(false);

    const tryFetch = useCallback(async ({ status, relevancyScore, subReddit, dateRange, pageCount = DEFAULT_DATA_LIMIT, pageNo = 1 }: FetchFilters) => {
        return await portalClient.getRelevantLeads({
            ...(relevancyScore && { relevancyScore }),
            ...(subReddit && { subReddit }),
            ...(status && { status }),
            ...(dateRange && { dateRange }),
            pageCount,
            pageNo,
        });
    }, []);

    const fetchLeads = useCallback(async ({ pageNo = 1, append = false, usePriority = false, loadType = "initial" }: fetchLeadsProps) => {
        try {
            if (loadType === "initial") {
                dispatch(setIsLoading(true));
            } else {
                setIsFetchingMore(true);
            }

            if (usePriority) {
                const leadStatusPriority: LeadStatus[] = [LeadStatus.NEW, LeadStatus.COMPLETED];

                for (const status of leadStatusPriority) {
                    try {
                        const result = await tryFetch({
                            status,
                            relevancyScore,
                            subReddit,
                            dateRange,
                            pageNo,
                        });
                        const leads = result?.leads ?? [];

                        if (leads.length > 0) {
                            dispatch(setNewTabList(leads));
                            dispatch(setLeadStatusFilter(status));
                            setCounts?.(result.analysis);
                            setHasMore(leads.length >= DEFAULT_DATA_LIMIT - 1);
                            setPageNo(1);
                            break;
                        }
                    } catch (error) {
                        console.error(`Error fetching leads for status: ${status}`, error);
                    }
                }

                hasCheckedInitialLeads.current = true;
            } else {
                const result = await tryFetch({
                    status: leadStatusFilter,
                    relevancyScore,
                    subReddit,
                    dateRange,
                    pageNo,
                });

                const leads = result.leads ?? [];

                if (append) {
                    dispatch(setNewTabList([...newTabList, ...leads]));
                } else {
                    dispatch(setNewTabList(leads));
                }

                setCounts?.(result.analysis);
                setHasMore(leads.length >= DEFAULT_DATA_LIMIT - 1);
                setPageNo(pageNo);
            }
        } catch (err: any) {
            const message =
                err?.response?.data?.message || err.message || "Something went wrong";
            toast({ title: "Error", description: message });
            dispatch(setError(message));
        } finally {
            if (loadType === "initial") {
                dispatch(setIsLoading(false));
            } else {
                setIsFetchingMore(false);
            }
        }
    }, [dispatch, setIsFetchingMore, tryFetch, relevancyScore, subReddit, dateRange, setCounts, setHasMore, setPageNo, leadStatusFilter, newTabList]);

    return { fetchLeads };
};