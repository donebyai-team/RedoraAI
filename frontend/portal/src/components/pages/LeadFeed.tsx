"use client";

import { useEffect, useState } from "react";
// import { toast } from "@/components/ui/use-toast";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";
import { LeadFeed as LeadFeedComponent } from "@/components/dashboard/LeadFeed";
import { RelevancyScoreSidebar } from "@/components/dashboard/RelevancyScoreSidebar";
import { FilterControls } from "@/components/dashboard/FilterControls";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setError, setIsLoading, setLeadStatusFilter, setNewTabList } from "@/store/Lead/leadSlice";
import { toast } from "@/components/ui/use-toast";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { DEFAULT_DATA_LIMIT, FetchFilters } from "./Dashboard";

export default function LeadFeed() {

  const { portalClient } = useClientsContext()
  const dispatch = useAppDispatch();
  const { dateRange, leadStatusFilter, isLoading, newTabList } = useAppSelector((state) => state.lead);
  const { relevancyScore, subReddit } = useAppSelector((state) => state.parems);
  const [pageNo, setPageNo] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);

  const tryFetch = async ({ status, relevancyScore, subReddit, dateRange, pageCount = DEFAULT_DATA_LIMIT, pageNo = 0 }: FetchFilters) => {
    const result = await portalClient.getRelevantLeads({
      ...(relevancyScore && { relevancyScore }),
      ...(subReddit && { subReddit }),
      ...(status && { status }),
      ...(dateRange && { dateRange }),
      pageCount,
      pageNo
    });

    return result;
  };

  const fetchLeads = async ({
    page = 1,
    append = false,
    loadType = "initial", // "initial" | "pagination"
  }: {
    page?: number;
    append?: boolean;
    loadType?: "initial" | "pagination";
  }) => {
    try {
      if (loadType === "initial") {
        dispatch(setIsLoading(true));
      } else {
        setIsFetchingMore(true);
      }

      const result = await tryFetch({
        status: leadStatusFilter,
        relevancyScore,
        subReddit,
        dateRange,
        pageNo: page,
        pageCount: DEFAULT_DATA_LIMIT,
      });

      const leads = result.leads ?? [];

      if (append) {
        dispatch(setNewTabList([...newTabList, ...leads]));
      } else {
        dispatch(setNewTabList(leads));
      }

      setHasMore(leads.length >= (DEFAULT_DATA_LIMIT - 1));
      setPageNo(page);

    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong";
      toast({ title: "Error", description: message });
      dispatch(setError(message));
    } finally {
      if (loadType === "initial") {
        dispatch(setIsLoading(false));
      } else {
        setIsFetchingMore(false);
      }
    }
  };

  useEffect(() => {

    fetchLeads({ page: 0, loadType: "initial" });

    setPageNo(0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dateRange, relevancyScore, subReddit, leadStatusFilter]);

  const loadMoreLeads = async () => {
    if (isFetchingMore || !hasMore) return;

    await fetchLeads({ page: pageNo + 1, append: true, loadType: "pagination" });
  };

  const renderTabContent = () => {
    if (isLoading) {
      return <div className="space-y-4">
        {[...Array(3)].map((_, i) => <Card key={i} className="border-primary/10 shadow-md">
          <CardContent className="p-6">
            <div className="space-y-2">
              <Skeleton className="h-4 w-[200px]" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-[80%]" />
              <div className="flex gap-2 pt-2">
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-9 w-20" />
              </div>
            </div>
          </CardContent>
        </Card>)}
      </div>;
    }
    return <div className="flex-1">
      <LeadFeedComponent
        loadMoreLeads={loadMoreLeads}
        hasMore={hasMore}
        isFetchingMore={isFetchingMore}
      />
    </div>;
  };

  const tabOptions: { label: string; status: LeadStatus }[] = [
    { label: "New", status: LeadStatus.NEW },
    { label: "Responded", status: LeadStatus.COMPLETED },
    { label: "Skipped", status: LeadStatus.NOT_RELEVANT },
    { label: "Saved", status: LeadStatus.LEAD },
  ];

  const handleLeadStatusFilterChange = (value: string) => {
    const selected = tabOptions.find((tab) => tab.label === value);
    dispatch(setLeadStatusFilter(selected?.status ?? null));
  };

  const activeTabLabel = tabOptions.find((tab) => tab.status === leadStatusFilter)?.label ?? tabOptions?.[0]?.label;

  return (
    <>
      <DashboardHeader />

      <div className="flex-1 overflow-auto">
        <main className="container mx-auto px-4 py-6 md:px-6">
          <div className="flex flex-col lg:flex-row gap-6">
            {/* Main content area */}
            <div className="flex-1 flex flex-col">
              <div className="space-y-2 mb-6">
                <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">Lead Feed</h1>
                <p className="text-muted-foreground">
                  Track and engage with potential leads from Reddit based on your keywords and subreddits.
                </p>
              </div>

              <div className="flex-1 flex flex-col space-y-4 mt-6">
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-background/95 py-2">
                  <h2 className="text-xl font-semibold">Relevant Posts</h2>
                  <FilterControls isLeadStatusFilter={false} />
                </div>

                {/* Tabs for filtering communications */}
                <Tabs value={activeTabLabel} onValueChange={handleLeadStatusFilterChange} className="w-full">
                  <TabsList className="grid w-full grid-cols-4 bg-secondary/50">
                    {tabOptions.map(({ label }) => (
                      <TabsTrigger
                        key={label}
                        value={label}
                        className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
                      >
                        {label}
                      </TabsTrigger>
                    ))}
                  </TabsList>

                  <div className="mt-4">
                    {renderTabContent()}
                  </div>
                </Tabs>
              </div>
            </div>

            {/* Sidebar */}
            <div className="lg:w-[300px] space-y-6">
              <RelevancyScoreSidebar />
            </div>
          </div>
        </main>
      </div>

      <DashboardFooter />
    </>
  );
}
