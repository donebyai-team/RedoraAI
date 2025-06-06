"use client";

import { useEffect, useRef, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  // ArrowUp,
  // MessageSquare,
  Search,
  // Bell,
  // User,
  // Filter,
  // Send,
  // Save,
  // X,
  Clock
} from "lucide-react";
import { toast } from "@/components/ui/use-toast";
import { Skeleton } from "@/components/ui/skeleton";
import { SummaryCards } from "@/components/dashboard/SummaryCards";
import { LeadFeed } from "@/components/dashboard/LeadFeed";
import { FilterControls } from "@/components/dashboard/FilterControls";
import { SidebarSettings } from "@/components/dashboard/SidebarSettings";
import { RelevancyScoreSidebar } from "@/components/dashboard/RelevancyScoreSidebar";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { RedditAccount } from "@/components/reddit-accounts/RedditAccountBadge";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { setError, setIsLoading, setLeadStatusFilter, setNewTabList } from "../../../store/Lead/leadSlice";
import { DateRangeFilter, LeadAnalysis } from "@doota/pb/doota/portal/v1/portal_pb";
import { setAccounts, setLoading } from "@/store/Reddit/RedditSlice";
import { useRedditIntegrationStatus } from "../Leads/Tabs/useRedditIntegrationStatus";
import { AnnouncementBanner } from "../dashboard/AnnouncementBanner";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";

interface FetchFilters {
  status: LeadStatus | null;
  relevancyScore?: number;
  subReddit?: string;
  dateRange?: DateRangeFilter;
  pageCount?: number;
}

export default function Dashboard() {
  const { portalClient } = useClientsContext()
  const dispatch = useAppDispatch();
  const project = useAppSelector((state) => state.stepper.project);
  const { dateRange, leadStatusFilter, isLoading } = useAppSelector((state) => state.lead);
  const { relevancyScore, subReddit } = useAppSelector((state) => state.parems);
  const { isConnected, loading: isLoadingRedditIntegrationStatus } = useRedditIntegrationStatus();

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

  const [counts, setCounts] = useState<LeadAnalysis | undefined>(undefined);
  const [defaultAccountId, setDefaultAccountId] = useState<string>("account1");

  const handleDefaultAccountChange = (accountId: string) => {
    setDefaultAccountId(accountId);
  };

  const hasCheckedInitialLeads = useRef(false);

  const tryFetch = async ({ status, relevancyScore, subReddit, dateRange, pageCount = 10, }: FetchFilters) => {
    const result = await portalClient.getRelevantLeads({
      ...(relevancyScore && { relevancyScore }),
      ...(subReddit && { subReddit }),
      ...(status && { status }),
      ...(dateRange && { dateRange }),
      pageCount,
    });

    return result;
  };

  useEffect(() => {

    const getInitialLeads = async () => {

      dispatch(setIsLoading(true));

      try {
        if (!hasCheckedInitialLeads.current) {
          const leadStatusPriority: LeadStatus[] = [LeadStatus.NEW, LeadStatus.COMPLETED];

          for (const status of leadStatusPriority) {
            const result = await tryFetch({ status, relevancyScore, dateRange });
            const leads = result.leads ?? [];

            if (leads.length > 0) {
              dispatch(setNewTabList(leads));
              dispatch(setLeadStatusFilter(status));
              setCounts(result.analysis);
              break;
            }

            if (!leads.length && status === LeadStatus.COMPLETED) {
              dispatch(setNewTabList([]));
              dispatch(setLeadStatusFilter(LeadStatus.NEW));
              setCounts(undefined);
            }
          }

          hasCheckedInitialLeads.current = true;
        } else {
          const response = await tryFetch({ status: leadStatusFilter, relevancyScore, subReddit, dateRange });
          dispatch(setNewTabList(response.leads ?? []));
          setCounts(response.analysis);
        }
      }
      catch (err: any) {
        const message = err?.response?.data?.message || err.message || "Something went wrong";
        toast({ title: "Error", description: message });
        dispatch(setError(message));
      } finally {
        dispatch(setIsLoading(false));
      }
    };

    if (isConnected) {
      getInitialLeads();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConnected, dateRange, relevancyScore, subReddit, leadStatusFilter, dispatch]);

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

  return (
    <>
      <DashboardHeader />
      {project && !project.isActive ? (
        <AnnouncementBanner message="⚠️ Your account has been paused due to inactivity. Please contact support to enable it." />
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

              <SummaryCards counts={counts} />

              {isLoadingRedditIntegrationStatus ? (<>
                Loading
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
                    // onAction={handleAction}
                    // redditAccounts={redditAccounts}
                    // defaultAccountId={defaultAccountId}
                    // onAccountChange={handlePostAccountChange}
                    />
                  </div>
                )}
              </div>
              )}
            </div>

            {/* Sidebar */}
            <div className="lg:w-[300px] space-y-6">
              <RelevancyScoreSidebar
                accounts={redditAccounts}
                defaultAccountId={defaultAccountId}
                onDefaultAccountChange={handleDefaultAccountChange}
              />

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
