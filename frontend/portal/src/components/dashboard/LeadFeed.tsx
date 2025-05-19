
import { useState, useEffect } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { MessageSquare, Save, Send, X, User, AlertTriangle } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { RedditAccount } from "@/components/reddit-accounts/RedditAccountBadge";
import { RedditAccountSelector } from "@/components/reddit-accounts/RedditAccountSelector";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "@/components/ui/use-toast";

interface LeadPost {
  id: string;
  title: string;
  snippet: string;
  subreddit: string;
  time: string;
  score: number;
  author: string;
  karma: string;
  tags: string[];
  aiSuggestion: string;
  aiDmSuggestion: string;
  assignedAccountId: string;
  lastReplied?: Date | null;
}

interface LeadFeedProps {
  onAction: (action: string, postId: string) => void;
  redditAccounts: RedditAccount[];
  defaultAccountId: string;
  onAccountChange: (postId: string, accountId: string) => void;
}

export function LeadFeed({ 
  onAction, 
  redditAccounts = [], 
  defaultAccountId,
  onAccountChange 
}: LeadFeedProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [recentlyUsedAccounts, setRecentlyUsedAccounts] = useState<Record<string, Date>>({});
  
  // Sample data with assigned Reddit accounts and last replied timestamp
  const posts: LeadPost[] = [
    {
      id: "post1",
      title: "Recommendations for SaaS email automation tools?",
      snippet: "We're a small B2B business looking to automate our email campaigns better. We've tried Mailchimp but it's not cutting it for more complex sequences...",
      subreddit: "r/SaaS",
      time: "2h ago",
      score: 0.94,
      author: "growthfounder",
      karma: "2.4k",
      tags: ["Recommendation", "Pain Point"],
      aiSuggestion: "I understand the frustration with basic email tools. For more complex B2B sequences, we've had great success with tools that specialize in behavior-triggered automation. Happy to share some specific recommendations that worked for our clients in the B2B space if you'd like.",
      aiDmSuggestion: "Hi there! Saw your post about email automation challenges. We've helped several B2B SaaS companies implement more effective email sequences. Would you be open to a quick chat about what specifically you're trying to achieve? I might be able to point you in the right direction.",
      assignedAccountId: redditAccounts.length > 0 ? redditAccounts[0].id : defaultAccountId,
      lastReplied: new Date(Date.now() - 1000 * 60 * 35) // 35 minutes ago
    },
    {
      id: "post2",
      title: "How do you find your first 100 B2B customers?",
      snippet: "We just launched our analytics platform for e-commerce, but finding early B2B customers is proving harder than expected...",
      subreddit: "r/startups",
      time: "5h ago",
      score: 0.89,
      author: "techfounder23",
      karma: "867",
      tags: ["Looking for Tools"],
      aiSuggestion: "For B2B SaaS, particularly in the analytics space, we've found content-led SEO to be incredibly effective for those first 100 customers. Creating highly targeted content that addresses specific pain points your platform solves can drive qualified leads. Would be happy to share some specific approaches that have worked for similar analytics platforms.",
      aiDmSuggestion: "Hi! I read your post about finding those first 100 B2B customers for your e-commerce analytics platform. This is exactly the challenge we help startups with. Would you be interested in hearing how we helped a similar analytics platform grow from 0 to 200+ clients in about 6 months using content-led SEO?",
      assignedAccountId: redditAccounts.length > 1 ? redditAccounts[1].id : defaultAccountId,
      lastReplied: null
    },
    {
      id: "post3",
      title: "LinkedIn outreach failing miserably - any alternatives?",
      snippet: "I've been trying to generate leads using LinkedIn for our SaaS product, but the response rate is abysmal...",
      subreddit: "r/Entrepreneur",
      time: "12h ago",
      score: 0.78,
      author: "marketingwiz",
      karma: "1.2k",
      tags: ["Pain Point"],
      aiSuggestion: "LinkedIn can be oversaturated these days. We've found Reddit to be an excellent alternative channel for B2B SaaS lead gen, particularly for products with a clearly defined audience. The key is genuine engagement rather than direct pitching. Happy to share some specific strategies that have worked well.",
      aiDmSuggestion: "Hi there! I noticed your post about LinkedIn outreach challenges. We've worked with several B2B SaaS companies to develop alternative channels (including Reddit!) that generated much better response rates. Would you be open to a quick chat about what might work better for your specific product?",
      assignedAccountId: defaultAccountId,
      lastReplied: new Date(Date.now() - 1000 * 60 * 10) // 10 minutes ago
    }
  ];

  // Simulate updating recently used accounts when an action is taken
  const updateRecentlyUsedAccounts = (postId: string) => {
    const post = posts.find(p => p.id === postId);
    if (!post) return;
    
    setRecentlyUsedAccounts(prev => ({
      ...prev,
      [post.assignedAccountId]: new Date()
    }));
  };
  
  // Handle action with recently used account tracking
  const handleAction = (action: string, postId: string) => {
    // Update recently used accounts
    if (action === 'comment' || action === 'dm') {
      updateRecentlyUsedAccounts(postId);
    }
    
    // Call the original onAction handler
    onAction(action, postId);
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.9) return "text-green-500 bg-green-50";
    if (score >= 0.7) return "text-amber-500 bg-amber-50";
    return "text-red-500 bg-red-50";
  };

  const toggleExpand = (id: string) => {
    setExpandedId(expandedId === id ? null : id);
  };

  const handleAccountChange = (postId: string, accountId: string) => {
    onAccountChange(postId, accountId);
  };

  // Function to check if the assigned account is valid for actions
  const isAccountValid = (postId: string) => {
    const post = posts.find(p => p.id === postId);
    if (!post) return false;
    
    const account = redditAccounts.find(acc => acc.id === post.assignedAccountId);
    if (!account) return false;
    
    return !account.status.isBanned && 
           !account.status.isFlagged && 
           (!account.status.cooldownMinutes || account.status.cooldownMinutes <= 0);
  };
  
  // Function to suggest alternative account if current one was recently used
  const shouldSuggestAccountRotation = (post: LeadPost) => {
    const currentAccount = redditAccounts.find(acc => acc.id === post.assignedAccountId);
    if (!currentAccount) return false;
    
    // If the account was used in the last hour
    const lastUsed = recentlyUsedAccounts[post.assignedAccountId];
    if (lastUsed && (new Date().getTime() - lastUsed.getTime()) < 60 * 60 * 1000) {
      // Find an available account that hasn't been used recently
      const availableAccount = redditAccounts.find(acc => 
        acc.id !== post.assignedAccountId && 
        acc.status.isActive && 
        !acc.status.isFlagged && 
        !acc.status.isBanned && 
        (!recentlyUsedAccounts[acc.id] || 
         (new Date().getTime() - recentlyUsedAccounts[acc.id].getTime()) > 60 * 60 * 1000)
      );
      
      return !!availableAccount;
    }
    
    return false;
  };
  
  // Show account rotation toast when post has a recently used account
  useEffect(() => {
    posts.forEach(post => {
      if (shouldSuggestAccountRotation(post)) {
        toast({
          title: "Consider rotating accounts",
          description: `The account u/${redditAccounts.find(acc => acc.id === post.assignedAccountId)?.username} was used recently. Consider using a different account for this post.`,
          duration: 5000,
        });
      }
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <ScrollArea className="h-[calc(100vh-300px)]">
      <div className="space-y-4 pr-4">
        {posts.map(post => (
          <Card key={post.id} className="overflow-hidden">
            <CardContent className="p-6">
              <div className="flex flex-col space-y-4">
                {/* Header with score and title */}
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-1">
                      <div className={`text-sm font-bold px-2 py-1 rounded-md ${getScoreColor(post.score)}`}>
                        {post.score.toFixed(2)}
                      </div>
                      <div className="flex gap-2">
                        {post.tags.map(tag => (
                          <span key={tag} className="text-xs bg-secondary px-2 py-0.5 rounded-full">
                            {tag}
                          </span>
                        ))}
                      </div>
                    </div>
                    <h3 className="text-lg font-medium">{post.title}</h3>
                    <div className="flex items-center text-sm text-muted-foreground gap-2 mt-1">
                      <span>{post.subreddit}</span>
                      <span>•</span>
                      <span>{post.time}</span>
                      <span>•</span>
                      <span>by {post.author}</span>
                      <span className="text-xs bg-secondary px-1.5 rounded-full">{post.karma} karma</span>
                    </div>
                  </div>
                </div>
                
                {/* Snippet and AI suggestion */}
                <p className="text-sm text-muted-foreground">{post.snippet}</p>
                
                <div className="bg-secondary/50 rounded-md p-3">
                  <div className="flex items-center gap-2 text-sm font-medium mb-2">
                    <MessageSquare className="h-4 w-4" />
                    <span>AI Suggested Response</span>
                  </div>
                  <p className="text-sm">{post.aiSuggestion}</p>
                </div>
                
                {/* Expanded DM suggestion */}
                {expandedId === post.id && (
                  <div className="bg-secondary/50 rounded-md p-3">
                    <div className="flex items-center gap-2 text-sm font-medium mb-2">
                      <Send className="h-4 w-4" />
                      <span>AI Suggested DM</span>
                    </div>
                    <p className="text-sm">{post.aiDmSuggestion}</p>
                  </div>
                )}
                
                {/* Reddit account selector with rotation suggestion if needed */}
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2 text-sm">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Replying as:</span>
                    
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div className="relative">
                            <RedditAccountSelector
                              accounts={redditAccounts}
                              currentAccountId={post.assignedAccountId || defaultAccountId}
                              onAccountChange={(accountId) => handleAccountChange(post.id, accountId)}
                              postId={post.id}
                            />
                            {shouldSuggestAccountRotation(post) && (
                              <div className="absolute -top-2 -right-2">
                                <div className="bg-amber-50 rounded-full p-0.5 border border-amber-200">
                                  <AlertTriangle className="h-3.5 w-3.5 text-amber-500" />
                                </div>
                              </div>
                            )}
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="top">
                          <p className="text-xs max-w-[200px]">
                            {shouldSuggestAccountRotation(post) 
                              ? "This account was used recently. Consider rotating accounts to avoid rate limits."
                              : "Using multiple Reddit accounts helps avoid rate limits and boosts reach."}
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  
                  {/* Show last replied time if available */}
                  {post.lastReplied && (
                    <span className="text-xs text-muted-foreground">
                      Last replied: {timeAgo(post.lastReplied)}
                    </span>
                  )}
                </div>
                
                {/* Action buttons */}
                <div className="flex flex-wrap gap-2 mt-2">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <div>
                          <button 
                            onClick={() => handleAction('comment', post.id)}
                            className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent disabled:opacity-50 disabled:pointer-events-none"
                            disabled={!isAccountValid(post.id)}
                          >
                            <MessageSquare className="h-4 w-4 mr-2" />
                            Post Comment
                          </button>
                        </div>
                      </TooltipTrigger>
                      {!isAccountValid(post.id) && (
                        <TooltipContent side="bottom">
                          <p className="text-xs">This account cannot be used for posting. Select an active account.</p>
                        </TooltipContent>
                      )}
                    </Tooltip>
                  </TooltipProvider>
                  
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <div>
                          <button 
                            onClick={() => handleAction('dm', post.id)}
                            className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent disabled:opacity-50 disabled:pointer-events-none"
                            disabled={!isAccountValid(post.id)}
                          >
                            <Send className="h-4 w-4 mr-2" />
                            Send DM
                          </button>
                        </div>
                      </TooltipTrigger>
                      {!isAccountValid(post.id) && (
                        <TooltipContent side="bottom">
                          <p className="text-xs">This account cannot be used for messaging. Select an active account.</p>
                        </TooltipContent>
                      )}
                    </Tooltip>
                  </TooltipProvider>
                  
                  <button 
                    onClick={() => handleAction('save', post.id)}
                    className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent"
                  >
                    <Save className="h-4 w-4 mr-2" />
                    Save
                  </button>
                  
                  <button 
                    onClick={() => handleAction('skip', post.id)}
                    className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent"
                  >
                    <X className="h-4 w-4 mr-2" />
                    Skip
                  </button>
                  
                  <button 
                    onClick={() => toggleExpand(post.id)}
                    className="ml-auto text-sm text-primary hover:underline"
                  >
                    {expandedId === post.id ? 'Hide DM suggestion' : 'Show DM suggestion'}
                  </button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </ScrollArea>
  );
}

// Helper function to format time ago
const timeAgo = (date: Date) => {
  const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);
  
  let interval = Math.floor(seconds / 31536000);
  if (interval >= 1) return `${interval} year${interval > 1 ? 's' : ''} ago`;
  
  interval = Math.floor(seconds / 2592000);
  if (interval >= 1) return `${interval} month${interval > 1 ? 's' : ''} ago`;
  
  interval = Math.floor(seconds / 86400);
  if (interval >= 1) return `${interval} day${interval > 1 ? 's' : ''} ago`;
  
  interval = Math.floor(seconds / 3600);
  if (interval >= 1) return `${interval} hour${interval > 1 ? 's' : ''} ago`;
  
  interval = Math.floor(seconds / 60);
  if (interval >= 1) return `${interval} minute${interval > 1 ? 's' : ''} ago`;
  
  return `${Math.floor(seconds)} second${seconds !== 1 ? 's' : ''} ago`;
};
