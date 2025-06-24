"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { toast } from "@/components/ui/use-toast";
import { PlusCircle, Tag, Trash2 } from "lucide-react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { portalClient } from "@/services/grpc";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setProject } from "@/store/Onboarding/OnboardingSlice";
import { Skeleton } from "@/components/ui/skeleton";

// Form validation schemas
const keywordSchema = z.object({
  keyword: z
    .string()
    .trim()
    .min(3, { message: "Keyword must be at least 3 characters" })
    .max(50, { message: "Keyword must be less than 50 characters" })
});

const subredditSchema = z.object({
  subreddit: z
    .string()
    .trim()
    .min(3, { message: "Subreddit must be at least 3 characters" })
    .max(21, { message: "Subreddit name can't exceed 21 characters" })
    .regex(/^[a-zA-Z0-9_]+$/, { message: "Subreddit can only contain letters, numbers, and underscores" })
});

const defaultKeywordLimit = 5;
const defaultSubredditLimit = 2;

export default function KeywordManagement() {

  const { getPlanDetails } = useAuth();
  const currentPlan = getPlanDetails();
  const project = useAppSelector((state) => state.stepper.project);
  const dispatch = useAppDispatch();

  const [suggestionLoading, setSuggestionLoading] = useState(true);
  const [isKeywordsLoading, setIsKeywordsLoading] = useState(false);
  const [isSubredditLoading, setIsSubredditLoading] = useState(false);

  // Plan limits
  const keywordLimit = currentPlan?.maxKeywords ?? defaultKeywordLimit;
  const subredditLimit = currentPlan?.maxSources ?? defaultSubredditLimit;

  const keywordForm = useForm<z.infer<typeof keywordSchema>>({
    resolver: zodResolver(keywordSchema),
    defaultValues: { keyword: "" },
  });

  const subredditForm = useForm<z.infer<typeof subredditSchema>>({
    resolver: zodResolver(subredditSchema),
    defaultValues: { subreddit: "" },
  });

  // Utilities
  const showToast = (title: string, description: string, variant: "default" | "destructive" = "default") =>
    toast({ title, description, variant });

  const hasReachedLimit = (count: number, limit: number, type: "Keyword" | "Subreddit") => {
    if (count >= limit) {
      showToast(`${type} limit reached`, `Your plan allows up to ${limit} ${type.toLowerCase()}s. Upgrade for more.`, "destructive");
      return true;
    }
    return false;
  };

  const isDuplicate = (list: string[], value: string) =>
    list.some(item => item.toLowerCase() === value.toLowerCase());

  // Suggestions
  const keywordSuggestions =
    project?.suggestedKeywords?.filter(suggestion =>
      !project.keywords.some(k => k.name.toLowerCase() === suggestion.toLowerCase())) ?? [];

  const subredditSuggestions =
    project?.suggestedSources?.filter(subreddit => {
      const plain = subreddit.replace(/^r\//i, "").toLowerCase();
      return !project.sources?.some(s => s.name.toLowerCase() === plain);
    }) ?? [];

  // Keyword Actions
  const addKeyword = async (keyword: string) => {
    if (!project) return;

    const names = project.keywords.map(k => k.name);
    if (hasReachedLimit(names.length, keywordLimit, "Keyword")) return;
    if (isDuplicate(names, keyword)) {
      showToast("Duplicate keyword", "This keyword is already being tracked.", "destructive");
    }

    setIsKeywordsLoading(true);
    try {
      const result = await portalClient.createKeywords({ keywords: [...names, keyword] });
      dispatch(setProject({ ...project, keywords: result.keywords }));
      showToast("Keyword added", `Now tracking "${keyword}"`);
    } catch (err: any) {
      showToast("Error", err?.response?.data?.message || err.message || "Something went wrong", "destructive");
    } finally {
      setIsKeywordsLoading(false);
    }
  };

  const handleAddKeyword = async (values: z.infer<typeof keywordSchema>) => {
    await addKeyword(values.keyword);
    keywordForm.reset();
  };

  const handleAddSuggestedKeyword = async (keyword: string) => {
    await addKeyword(keyword);
  };

  const handleDeleteKeyword = async (id: string) => {
    if (!project) return;
    setIsKeywordsLoading(true);

    try {
      const updatedKeywords = project.keywords.filter(k => k.id !== id).map(k => k.name);
      const result = await portalClient.createKeywords({ keywords: updatedKeywords });
      dispatch(setProject({ ...project, keywords: result.keywords }));
      showToast("Keyword removed", "The keyword has been removed from tracking.");
    } catch (err: any) {
      showToast("Error", err?.response?.data?.message || err.message || "Something went wrong", "destructive");
    } finally {
      setIsKeywordsLoading(false);
    }
  };

  // Subreddit Actions
  const addSubreddit = async (input: string) => {
    if (!project) return;

    const formatted = input.startsWith("r/") ? input : `r/${input}`;
    const names = project.sources.map(s => s.name);

    if (hasReachedLimit(names.length, subredditLimit, "Subreddit")) return;
    if (isDuplicate(names, formatted)) {
      showToast("Duplicate subreddit", "This subreddit is already being tracked.", "destructive");
    }

    setIsSubredditLoading(true);
    try {
      const result = await portalClient.addSource({ name: formatted });
      dispatch(setProject({ ...project, sources: [...project.sources, result] }));
      showToast("Subreddit added", `Now tracking ${formatted}`);
    } catch (err: any) {
      showToast("Error", err?.response?.data?.message || err.message || "Failed to add", "destructive");
    } finally {
      setIsSubredditLoading(false);
    }
  };

  const handleAddSubreddit = async (values: z.infer<typeof subredditSchema>) => {
    await addSubreddit(values.subreddit);
    subredditForm.reset();
  };

  const handleAddSuggestedSubreddit = async (subreddit: string) => {
    await addSubreddit(subreddit);
  };

  const handleDeleteSubreddit = async (id: string) => {
    if (!project) return;
    setIsSubredditLoading(true);

    try {
      await portalClient.removeSource({ id });
      const updated = project.sources.filter(s => s.id !== id);
      dispatch(setProject({ ...project, sources: updated }));
      showToast("Subreddit removed", "The subreddit has been removed from tracking.");
    } catch (err: any) {
      showToast("Error", err?.response?.data?.message || err.message || "Failed to remove", "destructive");
    } finally {
      setIsSubredditLoading(false);
    }
  };

  // Fetch Suggestions on Mount
  useEffect(() => {
    if (!project) return;
    (async () => {
      setSuggestionLoading(true);
      try {
        const result = await portalClient.suggestKeywordsAndSources({});
        dispatch(setProject(result));
      } catch (err: any) {
        console.error("Suggestion fetch failed:", err?.response?.data?.message || err.message);
      } finally {
        setSuggestionLoading(false);
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!project) return null;

  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-b from-background to-secondary/20">
      <DashboardHeader />

      <div className="flex-1 overflow-auto">
        <main className="container mx-auto px-4 py-6 md:px-6">
          <div className="space-y-2 mb-6">
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
              Keywords & Subreddits
            </h1>
            <p className="text-muted-foreground">
              Manage what you track to find the most relevant leads on Reddit.
            </p>
          </div>

          <div className="flex flex-col lg:flex-row gap-6">
            {/* Main content area */}
            <div className="flex-1">
              <Tabs defaultValue="keywords" className="space-y-4">
                <TabsList className="bg-secondary/50">
                  <TabsTrigger
                    className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
                    value="keywords"
                  >
                    Keywords
                  </TabsTrigger>
                  <TabsTrigger
                    className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
                    value="subreddits"
                  >
                    Subreddits
                  </TabsTrigger>
                </TabsList>

                {/* Keywords Tab */}
                <TabsContent value="keywords" className="space-y-4">
                  <Card className="border-primary/10 shadow-md">
                    <CardHeader>
                      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
                        <div>
                          <CardTitle className="flex items-center gap-2">
                            <Tag className="h-5 w-5" />
                            Tracked Keywords
                          </CardTitle>
                          <CardDescription>
                            Define the topics or buyer intent signals to track.
                          </CardDescription>
                        </div>
                        <Badge
                          variant={(project?.keywords?.length) >= keywordLimit ? "destructive" : "secondary"}
                          className="px-3 py-1"
                        >
                          {project?.keywords?.length}/{keywordLimit} Keywords Used
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-6">
                      {/* Add keyword form */}
                      <Form {...keywordForm}>
                        <form onSubmit={keywordForm.handleSubmit(handleAddKeyword)} className="flex flex-col sm:flex-row gap-2">
                          <FormField
                            control={keywordForm.control}
                            name="keyword"
                            disabled={isKeywordsLoading}
                            render={({ field }) => (
                              <FormItem className="flex-1">
                                <FormControl>
                                  <Input
                                    placeholder="Add a keyword to track (e.g., 'CRM for startups')"
                                    {...field}
                                  />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  type="submit"
                                  className="bg-primary gap-2"
                                  disabled={project?.keywords?.length >= keywordLimit || isKeywordsLoading}
                                >
                                  <PlusCircle className="h-4 w-4" />
                                  <span>Add Keyword</span>
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Add a short, high-intent phrase like 'tool for lead gen' or 'email warmup help'</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </form>
                      </Form>

                      {/* Keywords list */}
                      <ScrollArea className="h-[400px] pr-4">
                        <div className="space-y-3">
                          {project?.keywords?.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground italic">
                              No keywords added yet. Add your first keyword above.
                            </div>
                          ) : (
                            project.keywords?.map((keyword) => (
                              <div
                                key={keyword.id}
                                className="flex items-center justify-between p-3 border rounded-md hover:bg-secondary/50 transition-colors"
                              >
                                <div className="space-y-1">
                                  <p className="font-medium">"{keyword.name}"</p>
                                </div>
                                <div className="flex items-center gap-2">
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          onClick={() => handleDeleteKeyword(keyword.id)}
                                          disabled={isKeywordsLoading}
                                        >
                                          <Trash2 className="h-4 w-4 text-destructive" />
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>
                                        <p>Delete keyword</p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </div>
                              </div>
                            ))
                          )}
                        </div>
                      </ScrollArea>
                    </CardContent>
                  </Card>
                </TabsContent>

                {/* Subreddits Tab */}
                <TabsContent value="subreddits" className="space-y-4">
                  <Card className="border-primary/10 shadow-md">
                    <CardHeader>
                      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
                        <div>
                          <CardTitle className="flex items-center gap-2">
                            <Tag className="h-5 w-5" />
                            Tracked Subreddits
                          </CardTitle>
                          <CardDescription>
                            Choose communities where your potential customers discuss their needs.
                          </CardDescription>
                        </div>
                        <Badge
                          variant={project?.sources?.length >= subredditLimit ? "destructive" : "secondary"}
                          className="px-3 py-1"
                        >
                          {project?.sources?.length}/{subredditLimit} Subreddits Used
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-6">
                      {/* Add subreddit form */}
                      <Form {...subredditForm}>
                        <form onSubmit={subredditForm.handleSubmit(handleAddSubreddit)} className="flex flex-col sm:flex-row gap-2">
                          <FormField
                            control={subredditForm.control}
                            name="subreddit"
                            disabled={isSubredditLoading}
                            render={({ field }) => (
                              <FormItem className="flex-1">
                                <div className="relative">
                                  <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground">
                                    r/
                                  </div>
                                  <FormControl>
                                    <Input
                                      placeholder="marketing"
                                      {...field}
                                      className="pl-8"
                                      onChange={(e) => {
                                        // Remove r/ prefix if user types it
                                        const value = e.target.value.replace(/^r\//, '');
                                        field.onChange(value);
                                      }}
                                    />
                                  </FormControl>
                                </div>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  type="submit"
                                  className="bg-primary gap-2"
                                  disabled={project?.sources?.length >= subredditLimit || isSubredditLoading}
                                >
                                  <PlusCircle className="h-4 w-4" />
                                  <span>Add Subreddit</span>
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Add niche communities where your audience asks for help (e.g., r/saas, r/emailmarketing)</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </form>
                      </Form>

                      {/* Subreddits list */}
                      <ScrollArea className="h-[400px] pr-4">
                        <div className="space-y-3">
                          {project?.sources?.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground italic">
                              No subreddits added yet. Add your first subreddit above.
                            </div>
                          ) : (
                            project?.sources?.map((subreddit) => (
                              <div
                                key={subreddit.id}
                                className="flex items-center justify-between p-3 border rounded-md hover:bg-secondary/50 transition-colors"
                              >
                                <div className="space-y-1">
                                  <div className="flex items-center gap-2">
                                    <p className="font-medium">{subreddit.name}</p>
                                  </div>
                                </div>
                                <div className="flex items-center gap-2">
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          disabled={isSubredditLoading}
                                          onClick={() => handleDeleteSubreddit(subreddit.id)}
                                        >
                                          <Trash2 className="h-4 w-4 text-destructive" />
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>
                                        <p>Delete subreddit</p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </div>
                              </div>
                            ))
                          )}
                        </div>
                      </ScrollArea>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>

            {/* Suggestions Sidebar */}
            <div className="lg:w-[300px] space-y-6">
              <Card className="border-primary/10 shadow-md">
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <PlusCircle className="h-4 w-4" />
                    AI Suggestions
                  </CardTitle>
                  <CardDescription>
                    Personalized recommendations based on your tracking history.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <h4 className="font-medium text-sm">Suggested Keywords</h4>

                    {suggestionLoading ? (<>
                      {[...Array(3)].map((_, i) => (
                        <div key={i} className="flex rounded-md">
                          <Skeleton className="h-10 w-full" />
                        </div>
                      ))}
                    </>) : (<>
                      {keywordSuggestions.map((keyword, i) => (
                        <div key={i} className="flex justify-between items-center p-2 rounded-md hover:bg-secondary/50">
                          <p className="text-sm">"{keyword}"</p>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={isKeywordsLoading}
                            onClick={() => handleAddSuggestedKeyword(keyword)}
                          >
                            <PlusCircle className="h-4 w-4" />
                          </Button>
                        </div>
                      ))}
                    </>)}
                  </div>

                  <div className="space-y-4">
                    <h4 className="font-medium text-sm">Suggested Subreddits</h4>

                    {suggestionLoading ? (<>
                      {[...Array(3)].map((_, i) => (
                        <div key={i} className="flex rounded-md">
                          <Skeleton className="h-10 w-full" />
                        </div>
                      ))}
                    </>) : (<>
                      {subredditSuggestions.map((subreddit, i) => (
                        <div key={i} className="flex justify-between items-center p-2 rounded-md hover:bg-secondary/50">
                          <p className="text-sm">{subreddit}</p>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={isSubredditLoading}
                            onClick={() => handleAddSuggestedSubreddit(subreddit)}
                          >
                            <PlusCircle className="h-4 w-4" />
                          </Button>
                        </div>
                      ))}
                    </>)}
                  </div>

                  <div className="bg-primary/5 p-3 rounded-md">
                    <h4 className="font-medium text-sm mb-2">Insights</h4>
                    <p className="text-xs text-muted-foreground">
                      Your top-performing keyword is "lead generation" with 35 matches this week.
                    </p>
                    <p className="text-xs text-muted-foreground mt-2">
                      r/startups has generated the most leads for your business.
                    </p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </main>
      </div>

      <DashboardFooter />
    </div>
  );
}