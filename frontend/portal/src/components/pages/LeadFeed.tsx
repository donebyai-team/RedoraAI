"use client";

import { useState } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";
import { LeadFeed as LeadFeedComponent } from "@/components/dashboard/LeadFeed";
import { RelevancyScoreSidebar } from "@/components/dashboard/RelevancyScoreSidebar";
import { FilterControls } from "@/components/dashboard/FilterControls";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setLeadStatusFilter } from "@/store/Lead/leadSlice";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useLeadListManager } from "@/hooks/useLeadListManager";
import { useSetLeadFilters } from "@/hooks/useSetLeadFilters";

export default function LeadFeed() {

  const dispatch = useAppDispatch();
  const { dateRange, leadStatusFilter, isLoading, leadList } = useAppSelector((state) => state.lead);
  const { relevancyScore, subReddit } = useAppSelector((state) => state.parems);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);
  const { resetData } = useSetLeadFilters();

  const { loadMoreLeads } = useLeadListManager({
    relevancyScore,
    subReddit,
    dateRange,
    leadStatusFilter,
    leadList,
    setHasMore,
    setIsFetchingMore,
  });

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
    resetData();
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
              <div className="space-y-2 mb-2">
                <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">Tracked Conversations</h1>
                <p className="text-muted-foreground">
                  Highly relevant community discussions powered by your keywords and product capabilities.
                </p>
              </div>

              <div className="flex-1 flex flex-col space-y-4 mt-4">
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-background/95 py-2">
                  {/* <h2 className="text-xl font-semibold">Relevant Posts</h2> */}
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
