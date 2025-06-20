"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Search,
  Clock
} from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { SummaryCards } from "@/components/dashboard/SummaryCards";
import { LeadFeed } from "@/components/dashboard/LeadFeed";
import { FilterControls } from "@/components/dashboard/FilterControls";
import { SidebarSettings } from "@/components/dashboard/SidebarSettings";
import { RelevancyScoreSidebar } from "@/components/dashboard/RelevancyScoreSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { setError } from "../../../store/Lead/leadSlice";
import { setAccounts, setLoading } from "@/store/Reddit/RedditSlice";
import { useRedditIntegrationStatus } from "../Leads/Tabs/useRedditIntegrationStatus";
import { AnnouncementBanner } from "../dashboard/AnnouncementBanner";
import { SubscriptionStatus } from "@doota/pb/doota/core/v1/core_pb";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { useLeadListManager } from "@/hooks/useLeadListManager";

export default function Dashboard() {
  const { portalClient } = useClientsContext()
  const { currentOrganization } = useAuth()
  const dispatch = useAppDispatch();
  const project = useAppSelector((state) => state.stepper.project);
  const { dateRange, leadStatusFilter, isLoading, leadList, dashboardCounts } = useAppSelector((state) => state.lead);
  const { relevancyScore, subReddit } = useAppSelector((state) => state.parems);
  const { isConnected, loading: isLoadingRedditIntegrationStatus } = useRedditIntegrationStatus();
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);

  const { loadMoreLeads } = useLeadListManager({
    relevancyScore,
    subReddit,
    dateRange,
    leadStatusFilter,
    leadList,
    setHasMore,
    setIsFetchingMore,
  });

  console.log("####_123");

  // get all reddit account, used in Leed Feed
  useEffect(() => {
    dispatch(setLoading(true));
    portalClient.getIntegrations({})
      .then((res) => {
        dispatch(setAccounts(res.integrations));
      })
      .catch((err) => {
        dispatch(setError('Failed to fetch integrations'));
        console.error("Error fetching integrations:", err);
      })
      .finally(() => {
        dispatch(setLoading(false));
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // check user plan is expire ot not
  const isPLanExpire = (currentOrganization && currentOrganization?.featureFlags?.subscription?.status === SubscriptionStatus.EXPIRED);

  return (
    <>
      <DashboardHeader />
      {isPLanExpire ? (
        <AnnouncementBanner
          message="🚫 Your free trial has ended. Please upgrade your plan to continue using RedoraAI."
          buttonText="Upgrade now →"
          buttonHref="/settings/billing"
        />
      ) : project && !project.isActive ? (
        <AnnouncementBanner
          message="⚠️ Your account has been paused due to inactivity."
          buttonText="Reactivate now →"
          buttonHref="/settings/automation"
        />
      ) : (!isConnected && !isLoadingRedditIntegrationStatus) ? (
        <AnnouncementBanner
          message="⚠️ Connect your Reddit account to get real-time alerts and auto-reply to trending posts."
          buttonText="Connect now →"
          buttonHref="/settings/integrations"
        />
      ) : null}

      <div className="flex-1 overflow-auto">
        <main className="container mx-auto px-4 py-6 md:px-6">
          <div className="flex flex-col lg:flex-row gap-6">
            {/* Main content area */}
            <div className="flex-1 flex flex-col">
              <div className="space-y-2 mb-6">
                <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">Redora AI Dashboard</h1>
                <p className="text-muted-foreground">
                  Track and engage with potential leads from Reddit based on your keywords and subreddits.
                </p>
              </div>

              <SummaryCards counts={dashboardCounts} />

              {isLoadingRedditIntegrationStatus ? (<>
                <div className="flex-1 flex justify-center space-y-4 mt-6">
                  <h5 className="text-xl font-semibold">Loading...</h5>
                </div>
              </>) : (<div className="flex-1 flex flex-col space-y-4 mt-6">
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-background/95 py-2">
                  <h2 className="text-xl font-semibold">Latest Tracked Posts</h2>
                  <FilterControls />
                </div>

                {isLoading ? (
                  <div className="space-y-4">
                    {[...Array(3)].map((_, i) => (
                      <Card key={i} className="border-primary/10 shadow-md">
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
                      </Card>
                    ))}
                  </div>
                ) : (
                  <div className="flex-1">
                    <LeadFeed
                      loadMoreLeads={loadMoreLeads}
                      hasMore={hasMore}
                      isFetchingMore={isFetchingMore}
                    />
                  </div>
                )}
              </div>
              )}
            </div>

            {/* Sidebar */}
            <div className="lg:w-[300px] space-y-6">
              <RelevancyScoreSidebar />

              <Card className="border-primary/10 shadow-md">
                <CardContent className="p-6">
                  <Tabs defaultValue="keywords">
                    <TabsList className="w-full mb-4 bg-secondary/50">
                      <TabsTrigger className="flex-1 data-[state=active]:bg-primary/10 data-[state=active]:text-primary" value="keywords">Keywords</TabsTrigger>
                      <TabsTrigger className="flex-1 data-[state=active]:bg-primary/10 data-[state=active]:text-primary" value="subreddits">Subreddits</TabsTrigger>
                    </TabsList>
                    <TabsContent value="keywords" className="space-y-4">
                      <SidebarSettings type="keywords" />
                    </TabsContent>
                    <TabsContent value="subreddits" className="space-y-4">
                      <SidebarSettings type="subreddits" />
                    </TabsContent>
                  </Tabs>
                </CardContent>
              </Card>

              <Card className="border-primary/10 bg-gradient-to-br from-background to-secondary/30 shadow-md">
                <CardContent className="p-6">
                  <h3 className="text-lg font-medium mb-4">Tips</h3>
                  <div className="space-y-4 text-sm">
                    <div className="flex gap-2 items-start">
                      <div className="bg-primary/10 p-2 rounded-full">
                        <Search className="h-4 w-4 text-primary" />
                      </div>
                      <p>We score every post based on how well it matches your ideal customer and their pain points.</p>
                    </div>
                    <div className="flex gap-2 items-start">
                      <div className="bg-primary/10 p-2 rounded-full">
                        <Clock className="h-4 w-4 text-primary" />
                      </div>
                      <p>Redora scans Reddit 24/7 so you never miss a potential buyer conversation.</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </main>
      </div>

      <DashboardFooter />
    </>
  );
}
