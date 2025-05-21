
import { useState } from "react";
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
import { PlusCircle, X, Tag, Clock, BarChart2, Pin, PinOff, Edit, Trash2, Check } from "lucide-react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";

// Define types for our data
interface KeywordData {
  id: string;
  name: string;
  matches: number;
  lastSeen: string;
  subreddit?: string;
}

interface SubredditData {
  id: string;
  name: string;
  activity: number;
  lastMatched: string;
  matchedKeywords: string[];
  isPinned: boolean;
}

// Form validation schemas
const keywordSchema = z.object({
  keyword: z
    .string()
    .min(3, { message: "Keyword must be at least 3 characters" })
    .max(50, { message: "Keyword must be less than 50 characters" })
});

const subredditSchema = z.object({
  subreddit: z
    .string()
    .min(3, { message: "Subreddit must be at least 3 characters" })
    .max(21, { message: "Subreddit name can't exceed 21 characters" })
    .regex(/^[a-zA-Z0-9_]+$/, { message: "Subreddit can only contain letters, numbers, and underscores" })
});

export default function KeywordManagement() {
  // State for keywords and subreddits
  const [keywords, setKeywords] = useState<KeywordData[]>([
    { id: "k1", name: "lead generation", matches: 35, lastSeen: "3h ago", subreddit: "r/marketing" },
    { id: "k2", name: "email automation", matches: 28, lastSeen: "12h ago", subreddit: "r/startups" },
    { id: "k3", name: "CRM tool", matches: 12, lastSeen: "2h ago", subreddit: "r/SaaS" },
    { id: "k4", name: "SEO strategy", matches: 47, lastSeen: "1d ago", subreddit: "r/SEO" },
  ]);

  const [subreddits, setSubreddits] = useState<SubredditData[]>([
    { id: "s1", name: "r/startups", activity: 42, lastMatched: "3h ago", matchedKeywords: ["lead generation", "CRM tool"], isPinned: true },
    { id: "s2", name: "r/Entrepreneur", activity: 38, lastMatched: "12h ago", matchedKeywords: ["email automation"], isPinned: false },
    { id: "s3", name: "r/SaaS", activity: 27, lastMatched: "2h ago", matchedKeywords: ["CRM tool"], isPinned: true },
    { id: "s4", name: "r/marketing", activity: 19, lastMatched: "1d ago", matchedKeywords: ["SEO strategy", "lead generation"], isPinned: false },
  ]);

  const [editingKeywordId, setEditingKeywordId] = useState<string | null>(null);
  const [editingSubredditId, setEditingSubredditId] = useState<string | null>(null);

  // Plan limits
  const keywordLimit = 5; // Can be changed based on user's plan
  const subredditLimit = 10;

  // Forms for adding keywords and subreddits
  const keywordForm = useForm<z.infer<typeof keywordSchema>>({
    resolver: zodResolver(keywordSchema),
    defaultValues: {
      keyword: "",
    },
  });

  const subredditForm = useForm<z.infer<typeof subredditSchema>>({
    resolver: zodResolver(subredditSchema),
    defaultValues: {
      subreddit: "",
    },
  });

  // Suggestions from AI analysis
  const keywordSuggestions = ["competitive analysis", "outreach tools", "cold email"];
  const subredditSuggestions = ["r/sales", "r/digitalmarketing", "r/growmybusiness"];

  const handleAddKeyword = (values: z.infer<typeof keywordSchema>) => {
    if (keywords.length >= keywordLimit) {
      toast({
        title: "Keyword limit reached",
        description: `Your plan allows up to ${keywordLimit} keywords. Upgrade for more.`,
        variant: "destructive",
      });
      return;
    }

    // Check for duplicates
    if (keywords.some(k => k.name.toLowerCase() === values.keyword.toLowerCase())) {
      toast({
        title: "Duplicate keyword",
        description: "This keyword is already being tracked.",
        variant: "destructive",
      });
      return;
    }

    const newKeyword: KeywordData = {
      id: `k${Date.now()}`,
      name: values.keyword,
      matches: 0,
      lastSeen: "Just added",
    };

    setKeywords(prev => [...prev, newKeyword]);
    keywordForm.reset();

    toast({
      title: "Keyword added",
      description: `Now tracking "${values.keyword}"`,
    });
  };

  const handleAddSubreddit = (values: z.infer<typeof subredditSchema>) => {
    if (subreddits.length >= subredditLimit) {
      toast({
        title: "Subreddit limit reached",
        description: `Your plan allows up to ${subredditLimit} subreddits. Upgrade for more.`,
        variant: "destructive",
      });
      return;
    }

    const subredditName = values.subreddit.startsWith('r/')
      ? values.subreddit
      : `r/${values.subreddit}`;

    // Check for duplicates
    if (subreddits.some(s => s.name.toLowerCase() === subredditName.toLowerCase())) {
      toast({
        title: "Duplicate subreddit",
        description: "This subreddit is already being tracked.",
        variant: "destructive",
      });
      return;
    }

    // Here we would typically validate if this is a real subreddit via Reddit API
    // For now, we'll simulate that this check passed

    const newSubreddit: SubredditData = {
      id: `s${Date.now()}`,
      name: subredditName,
      activity: 0,
      lastMatched: "Just added",
      matchedKeywords: [],
      isPinned: false
    };

    setSubreddits(prev => [...prev, newSubreddit]);
    subredditForm.reset();

    toast({
      title: "Subreddit added",
      description: `Now tracking ${subredditName}`,
    });
  };

  const handleEditKeyword = (id: string, newName: string) => {
    if (keywords.some(k => k.id !== id && k.name.toLowerCase() === newName.toLowerCase())) {
      toast({
        title: "Duplicate keyword",
        description: "This keyword is already being tracked.",
        variant: "destructive",
      });
      return;
    }

    setKeywords(prev => prev.map(k =>
      k.id === id ? { ...k, name: newName } : k
    ));
    setEditingKeywordId(null);

    toast({
      title: "Keyword updated",
      description: `Updated to "${newName}"`,
    });
  };

  const handleEditSubreddit = (id: string, newName: string) => {
    const formattedName = newName.startsWith('r/') ? newName : `r/${newName}`;

    if (subreddits.some(s => s.id !== id && s.name.toLowerCase() === formattedName.toLowerCase())) {
      toast({
        title: "Duplicate subreddit",
        description: "This subreddit is already being tracked.",
        variant: "destructive",
      });
      return;
    }

    setSubreddits(prev => prev.map(s =>
      s.id === id ? { ...s, name: formattedName } : s
    ));
    setEditingSubredditId(null);

    toast({
      title: "Subreddit updated",
      description: `Updated to ${formattedName}`,
    });
  };

  const handleDeleteKeyword = (id: string) => {
    setKeywords(prev => prev.filter(k => k.id !== id));
    toast({
      title: "Keyword removed",
      description: "The keyword has been removed from tracking.",
    });
  };

  const handleDeleteSubreddit = (id: string) => {
    setSubreddits(prev => prev.filter(s => s.id !== id));
    toast({
      title: "Subreddit removed",
      description: "The subreddit has been removed from tracking.",
    });
  };

  const handleTogglePin = (id: string) => {
    setSubreddits(prev => prev.map(s =>
      s.id === id ? { ...s, isPinned: !s.isPinned } : s
    ));
  };

  const handleAddSuggestedKeyword = (keyword: string) => {
    if (keywords.length >= keywordLimit) {
      toast({
        title: "Keyword limit reached",
        description: `Your plan allows up to ${keywordLimit} keywords. Upgrade for more.`,
        variant: "destructive",
      });
      return;
    }

    const newKeyword: KeywordData = {
      id: `k${Date.now()}`,
      name: keyword,
      matches: 0,
      lastSeen: "Just added",
    };

    setKeywords(prev => [...prev, newKeyword]);

    toast({
      title: "Suggested keyword added",
      description: `Now tracking "${keyword}"`,
    });
  };

  const handleAddSuggestedSubreddit = (subreddit: string) => {
    if (subreddits.length >= subredditLimit) {
      toast({
        title: "Subreddit limit reached",
        description: `Your plan allows up to ${subredditLimit} subreddits. Upgrade for more.`,
        variant: "destructive",
      });
      return;
    }

    const newSubreddit: SubredditData = {
      id: `s${Date.now()}`,
      name: subreddit,
      activity: 0,
      lastMatched: "Just added",
      matchedKeywords: [],
      isPinned: false
    };

    setSubreddits(prev => [...prev, newSubreddit]);

    toast({
      title: "Suggested subreddit added",
      description: `Now tracking ${subreddit}`,
    });
  };

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
                          variant={keywords.length >= keywordLimit ? "destructive" : "secondary"}
                          className="px-3 py-1"
                        >
                          {keywords.length}/{keywordLimit} Keywords Used
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
                                  disabled={keywords.length >= keywordLimit}
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
                          {keywords.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground italic">
                              No keywords added yet. Add your first keyword above.
                            </div>
                          ) : (
                            keywords.map((keyword) => (
                              <div
                                key={keyword.id}
                                className="flex items-center justify-between p-3 border rounded-md hover:bg-secondary/50 transition-colors"
                              >
                                <div className="space-y-1">
                                  {editingKeywordId === keyword.id ? (
                                    <div className="flex gap-2">
                                      <Input
                                        defaultValue={keyword.name}
                                        className="w-56 h-8"
                                        onKeyDown={(e) => {
                                          if (e.key === 'Enter') {
                                            handleEditKeyword(keyword.id, e.currentTarget.value);
                                          } else if (e.key === 'Escape') {
                                            setEditingKeywordId(null);
                                          }
                                        }}
                                      />
                                      <Button
                                        size="sm"
                                        variant="secondary"
                                        onClick={() => setEditingKeywordId(null)}
                                      >
                                        <X className="h-4 w-4" />
                                      </Button>
                                      <Button
                                        size="sm"
                                        onClick={(e) => {
                                          const input = e.currentTarget.previousSibling?.previousSibling as HTMLInputElement;
                                          handleEditKeyword(keyword.id, input.value);
                                        }}
                                      >
                                        <Check className="h-4 w-4" />
                                      </Button>
                                    </div>
                                  ) : (
                                    <p className="font-medium">"{keyword.name}"</p>
                                  )}
                                  <div className="flex flex-wrap gap-2 text-xs">
                                    <div className="flex items-center gap-1 text-muted-foreground">
                                      <Clock className="h-3 w-3" />
                                      <span>Last seen {keyword.lastSeen}{keyword.subreddit && ` in ${keyword.subreddit}`}</span>
                                    </div>
                                    <div className="flex items-center gap-1 text-muted-foreground">
                                      <BarChart2 className="h-3 w-3" />
                                      <span>{keyword.matches} matches this week</span>
                                    </div>
                                  </div>
                                </div>
                                <div className="flex items-center gap-2">
                                  {editingKeywordId !== keyword.id && (
                                    <>
                                      <TooltipProvider>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Button
                                              variant="ghost"
                                              size="sm"
                                              onClick={() => setEditingKeywordId(keyword.id)}
                                            >
                                              <Edit className="h-4 w-4 text-muted-foreground" />
                                            </Button>
                                          </TooltipTrigger>
                                          <TooltipContent>
                                            <p>Edit keyword</p>
                                          </TooltipContent>
                                        </Tooltip>
                                      </TooltipProvider>
                                      <TooltipProvider>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Button
                                              variant="ghost"
                                              size="sm"
                                              onClick={() => handleDeleteKeyword(keyword.id)}
                                            >
                                              <Trash2 className="h-4 w-4 text-destructive" />
                                            </Button>
                                          </TooltipTrigger>
                                          <TooltipContent>
                                            <p>Delete keyword</p>
                                          </TooltipContent>
                                        </Tooltip>
                                      </TooltipProvider>
                                    </>
                                  )}
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
                          variant={subreddits.length >= subredditLimit ? "destructive" : "secondary"}
                          className="px-3 py-1"
                        >
                          {subreddits.length}/{subredditLimit} Subreddits Used
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
                                  disabled={subreddits.length >= subredditLimit}
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
                          {subreddits.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground italic">
                              No subreddits added yet. Add your first subreddit above.
                            </div>
                          ) : (
                            subreddits.map((subreddit) => (
                              <div
                                key={subreddit.id}
                                className="flex items-center justify-between p-3 border rounded-md hover:bg-secondary/50 transition-colors"
                              >
                                <div className="space-y-1">
                                  {editingSubredditId === subreddit.id ? (
                                    <div className="flex gap-2">
                                      <div className="relative">
                                        <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground">
                                          r/
                                        </div>
                                        <Input
                                          defaultValue={subreddit.name.replace(/^r\//, '')}
                                          className="w-56 h-8 pl-8"
                                          onKeyDown={(e) => {
                                            if (e.key === 'Enter') {
                                              handleEditSubreddit(subreddit.id, e.currentTarget.value);
                                            } else if (e.key === 'Escape') {
                                              setEditingSubredditId(null);
                                            }
                                          }}
                                        />
                                      </div>
                                      <Button
                                        size="sm"
                                        variant="secondary"
                                        onClick={() => setEditingSubredditId(null)}
                                      >
                                        <X className="h-4 w-4" />
                                      </Button>
                                      <Button
                                        size="sm"
                                        onClick={(e) => {
                                          // Find the parent div that contains our input and buttons
                                          const parentDiv = e.currentTarget.closest('div');
                                          if (parentDiv) {
                                            // Find the input element inside the parent div
                                            const input = parentDiv.querySelector('input');
                                            if (input) {
                                              handleEditSubreddit(subreddit.id, input.value);
                                            }
                                          }
                                        }}
                                      >
                                        <Check className="h-4 w-4" />
                                      </Button>
                                    </div>
                                  ) : (
                                    <div className="flex items-center gap-2">
                                      <p className="font-medium">{subreddit.name}</p>
                                      {subreddit.isPinned && (
                                        <Badge variant="secondary" className="text-primary">Pinned</Badge>
                                      )}
                                    </div>
                                  )}
                                  <div className="flex flex-wrap gap-2 text-xs">
                                    <div className="flex items-center gap-1 text-muted-foreground">
                                      <Clock className="h-3 w-3" />
                                      <span>Last lead: {subreddit.lastMatched}</span>
                                    </div>
                                    <div className="flex items-center gap-1 text-muted-foreground">
                                      <BarChart2 className="h-3 w-3" />
                                      <span>{subreddit.activity} posts tracked this week</span>
                                    </div>
                                  </div>
                                  {subreddit.matchedKeywords.length > 0 && (
                                    <div className="flex flex-wrap gap-1 mt-1">
                                      {subreddit.matchedKeywords.map((keyword, i) => (
                                        <Badge key={i} variant="outline" className="text-xs">
                                          {keyword}
                                        </Badge>
                                      ))}
                                    </div>
                                  )}
                                </div>
                                <div className="flex items-center gap-2">
                                  {editingSubredditId !== subreddit.id && (
                                    <>
                                      <TooltipProvider>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Button
                                              variant="ghost"
                                              size="sm"
                                              onClick={() => handleTogglePin(subreddit.id)}
                                            >
                                              {subreddit.isPinned ? (
                                                <PinOff className="h-4 w-4 text-muted-foreground" />
                                              ) : (
                                                <Pin className="h-4 w-4 text-muted-foreground" />
                                              )}
                                            </Button>
                                          </TooltipTrigger>
                                          <TooltipContent>
                                            <p>{subreddit.isPinned ? "Unpin subreddit" : "Pin subreddit"}</p>
                                          </TooltipContent>
                                        </Tooltip>
                                      </TooltipProvider>
                                      <TooltipProvider>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Button
                                              variant="ghost"
                                              size="sm"
                                              onClick={() => setEditingSubredditId(subreddit.id)}
                                            >
                                              <Edit className="h-4 w-4 text-muted-foreground" />
                                            </Button>
                                          </TooltipTrigger>
                                          <TooltipContent>
                                            <p>Edit subreddit</p>
                                          </TooltipContent>
                                        </Tooltip>
                                      </TooltipProvider>
                                      <TooltipProvider>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Button
                                              variant="ghost"
                                              size="sm"
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
                                    </>
                                  )}
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
                    {keywordSuggestions.map((keyword, i) => (
                      <div key={i} className="flex justify-between items-center p-2 rounded-md hover:bg-secondary/50">
                        <p className="text-sm">"{keyword}"</p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleAddSuggestedKeyword(keyword)}
                        >
                          <PlusCircle className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>

                  <div className="space-y-4">
                    <h4 className="font-medium text-sm">Suggested Subreddits</h4>
                    {subredditSuggestions.map((subreddit, i) => (
                      <div key={i} className="flex justify-between items-center p-2 rounded-md hover:bg-secondary/50">
                        <p className="text-sm">{subreddit}</p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleAddSuggestedSubreddit(subreddit)}
                        >
                          <PlusCircle className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
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
