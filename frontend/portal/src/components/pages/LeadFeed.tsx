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
import { RedditAccount } from "@/components/reddit-accounts/RedditAccountBadge";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setError, setIsLoading, setLeadStatusFilter, setNewTabList } from "@/store/Lead/leadSlice";
import { toast } from "@/components/ui/use-toast";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";

export default function LeadFeed() {

  const { portalClient } = useClientsContext()
  const dispatch = useAppDispatch();
  const { dateRange, leadStatusFilter, isLoading } = useAppSelector((state) => state.lead);
  const { relevancyScore, subReddit } = useAppSelector((state) => state.parems);
  const [activeTab, setActiveTab] = useState("All");

  // Sample Reddit accounts
  const redditAccounts: RedditAccount[] = [
    {
      id: "account1",
      username: "redora_official",
      karma: 2345,
      status: { isActive: true },
      isDefault: true
    },
    {
      id: "account2",
      username: "saas_helper",
      karma: 986,
      status: { isActive: true }
    },
    {
      id: "account3",
      username: "marketing_pro",
      karma: 75,
      status: { isActive: true, hasLowKarma: true }
    },
    {
      id: "account4",
      username: "startup_advisor",
      karma: 542,
      status: { isActive: false, cooldownMinutes: 35 }
    },
    {
      id: "account5",
      username: "b2b_expert",
      karma: 1203,
      status: { isActive: false, isFlagged: true }
    },
  ];

  const [defaultAccountId, setDefaultAccountId] = useState<string>("account1");

  const handleDefaultAccountChange = (accountId: string) => {
    setDefaultAccountId(accountId);
  };

  useEffect(() => {

    const getAllRelevantLeads = async () => {
      dispatch(setIsLoading(true));

      try {
        const result = await portalClient.getRelevantLeads({
          ...(relevancyScore && { relevancyScore }),
          ...(subReddit && { subReddit }),
          ...(leadStatusFilter && { status: leadStatusFilter }),
          dateRange,
          pageCount: 10
        });
        const allLeads = result.leads ?? [];
        dispatch(setNewTabList(allLeads));

      } catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong";
        toast({
          title: "Error",
          description: message,
        });
        dispatch(setError(message));
      } finally {
        dispatch(setIsLoading(false));
      }
    };

    getAllRelevantLeads();

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [relevancyScore, subReddit, dateRange, leadStatusFilter]);

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
      <LeadFeedComponent />
    </div>;
  };

  const handleLeadStatusFilterChange = (value: string) => {
    setActiveTab(value);
    const map: Record<string, LeadStatus> = {
      "All": 0,
      "Responded": LeadStatus.COMPLETED,
      "Skipped": LeadStatus.NOT_RELEVANT,
      "Saved": LeadStatus.LEAD,
    };

    dispatch(setLeadStatusFilter(map[value] ?? null));
  };

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
                  <h2 className="text-xl font-semibold">Active Leads</h2>
                  <FilterControls isLeadStatusFilter={false} />
                </div>

                {/* Tabs for filtering communications */}
                <Tabs value={activeTab} onValueChange={handleLeadStatusFilterChange} className="w-full">
                  <TabsList className="grid w-full grid-cols-4 bg-secondary/50">
                    <TabsTrigger value="All" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">All</TabsTrigger>
                    <TabsTrigger value="Responded" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">Responded</TabsTrigger>
                    <TabsTrigger value="Skipped" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">Skipped</TabsTrigger>
                    <TabsTrigger value="Saved" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">Saved</TabsTrigger>
                  </TabsList>

                  <div className="mt-4">
                    {renderTabContent()}
                  </div>
                </Tabs>
              </div>
            </div>

            {/* Sidebar */}
            <div className="lg:w-[300px] space-y-6">
              <RelevancyScoreSidebar
                accounts={redditAccounts}
                defaultAccountId={defaultAccountId}
                onDefaultAccountChange={handleDefaultAccountChange}
              />
            </div>
          </div>
        </main>
      </div>

      <DashboardFooter />
    </>
  );
}
