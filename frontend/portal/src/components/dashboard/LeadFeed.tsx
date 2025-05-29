
import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import {
  Brain,
  MessageSquare,
  Save,
  Send,
  ThumbsUp,
  // X,
  User,
  X,
  // AlertTriangle 
} from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
// import { RedditAccount } from "@/components/reddit-accounts/RedditAccountBadge";
// import { RedditAccountSelector } from "@/components/reddit-accounts/RedditAccountSelector";
import {
  Tooltip,
  TooltipContent,
  // TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "@/components/ui/use-toast";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { RootState } from "@/store/store";
import { getFormattedDate } from "@/utils/format";
import Link from "next/link";
import { HtmlBodyRenderer, HtmlTitleRenderer, MarkdownRenderer } from "../Html/HtmlRenderer";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { portalClient } from "@/services/grpc";
import { setNewTabList } from "@/store/Lead/leadSlice";

const tabNameMap: Record<LeadStatus, string> = {
  [LeadStatus.NEW]: "All",
  [LeadStatus.COMPLETED]: "Responded",
  [LeadStatus.NOT_RELEVANT]: "Skipped",
  [LeadStatus.LEAD]: "Saved",
};

export function getTabName(value: LeadStatus | null | undefined): string {
  if (!value) {
    return "All";
  }
  return tabNameMap[value];
}

export function LeadFeed() {

  // const [expandedId, setExpandedId] = useState<string | null>(null);
  const [selectedLeadId, setSelectedLeadId] = useState<string>("");
  // const [recentlyUsedAccounts, setRecentlyUsedAccounts] = useState<Record<string, Date>>({});
  const { newTabList, leadStatusFilter } = useAppSelector((state: RootState) => state.lead);
  const { subredditList } = useAppSelector((state: RootState) => state.source);
  const { accounts: redditAccounts } = useAppSelector((state) => state.reddit);
  const dispatch = useAppDispatch();

  // // Sample data with assigned Reddit accounts and last replied timestamp
  // const posts: LeadPost[] = [
  //   {
  //     id: "post1",
  //     title: "Recommendations for SaaS email automation tools?",
  //     snippet: "We're a small B2B business looking to automate our email campaigns better. We've tried Mailchimp but it's not cutting it for more complex sequences...",
  //     subreddit: "r/SaaS",
  //     time: "2h ago",
  //     score: 0.94,
  //     author: "growthfounder",
  //     karma: "2.4k",
  //     tags: ["Recommendation", "Pain Point"],
  //     aiSuggestion: "I understand the frustration with basic email tools. For more complex B2B sequences, we've had great success with tools that specialize in behavior-triggered automation. Happy to share some specific recommendations that worked for our clients in the B2B space if you'd like.",
  //     aiDmSuggestion: "Hi there! Saw your post about email automation challenges. We've helped several B2B SaaS companies implement more effective email sequences. Would you be open to a quick chat about what specifically you're trying to achieve? I might be able to point you in the right direction.",
  //     assignedAccountId: redditAccounts.length > 0 ? redditAccounts[0].id : defaultAccountId,
  //     lastReplied: new Date(Date.now() - 1000 * 60 * 35) // 35 minutes ago
  //   },
  //   {
  //     id: "post2",
  //     title: "How do you find your first 100 B2B customers?",
  //     snippet: "We just launched our analytics platform for e-commerce, but finding early B2B customers is proving harder than expected...",
  //     subreddit: "r/startups",
  //     time: "5h ago",
  //     score: 0.89,
  //     author: "techfounder23",
  //     karma: "867",
  //     tags: ["Looking for Tools"],
  //     aiSuggestion: "For B2B SaaS, particularly in the analytics space, we've found content-led SEO to be incredibly effective for those first 100 customers. Creating highly targeted content that addresses specific pain points your platform solves can drive qualified leads. Would be happy to share some specific approaches that have worked for similar analytics platforms.",
  //     aiDmSuggestion: "Hi! I read your post about finding those first 100 B2B customers for your e-commerce analytics platform. This is exactly the challenge we help startups with. Would you be interested in hearing how we helped a similar analytics platform grow from 0 to 200+ clients in about 6 months using content-led SEO?",
  //     assignedAccountId: redditAccounts.length > 1 ? redditAccounts[1].id : defaultAccountId,
  //     lastReplied: null
  //   },
  //   {
  //     id: "post3",
  //     title: "LinkedIn outreach failing miserably - any alternatives?",
  //     snippet: "I've been trying to generate leads using LinkedIn for our SaaS product, but the response rate is abysmal...",
  //     subreddit: "r/Entrepreneur",
  //     time: "12h ago",
  //     score: 0.78,
  //     author: "marketingwiz",
  //     karma: "1.2k",
  //     tags: ["Pain Point"],
  //     aiSuggestion: "LinkedIn can be oversaturated these days. We've found Reddit to be an excellent alternative channel for B2B SaaS lead gen, particularly for products with a clearly defined audience. The key is genuine engagement rather than direct pitching. Happy to share some specific strategies that have worked well.",
  //     aiDmSuggestion: "Hi there! I noticed your post about LinkedIn outreach challenges. We've worked with several B2B SaaS companies to develop alternative channels (including Reddit!) that generated much better response rates. Would you be open to a quick chat about what might work better for your specific product?",
  //     assignedAccountId: defaultAccountId,
  //     lastReplied: new Date(Date.now() - 1000 * 60 * 10) // 10 minutes ago
  //   }
  // ];

  const getScoreColor = (score: number) => {
    if (score >= 90) return "text-green-500 bg-green-50";
    if (score >= 70) return "text-amber-500 bg-amber-50";
    return "text-red-500 bg-red-50";
  };

  // const toggleExpand = (id: string) => {
  //   setExpandedId(expandedId === id ? null : id);
  // };

  // // Function to suggest alternative account if current one was recently used
  // const shouldSuggestAccountRotation = (post: LeadPost) => {
  //   const currentAccount = redditAccounts.find(acc => acc.id === post.assignedAccountId);
  //   if (!currentAccount) return false;

  //   // If the account was used in the last hour
  //   const lastUsed = recentlyUsedAccounts[post.assignedAccountId];
  //   if (lastUsed && (new Date().getTime() - lastUsed.getTime()) < 60 * 60 * 1000) {
  //     // Find an available account that hasn't been used recently
  //     const availableAccount = redditAccounts.find(acc =>
  //       acc.id !== post.assignedAccountId &&
  //       acc.status.isActive &&
  //       !acc.status.isFlagged &&
  //       !acc.status.isBanned &&
  //       (!recentlyUsedAccounts[acc.id] ||
  //         (new Date().getTime() - recentlyUsedAccounts[acc.id].getTime()) > 60 * 60 * 1000)
  //     );

  //     return !!availableAccount;
  //   }

  //   return false;
  // };

  const copyTextAndOpenLink = (textToCopy: string, linkToOpen: string) => {
    if (!navigator.clipboard) {
      // Fallback for older browsers that do not support `navigator.clipboard`
      const textArea = document.createElement("textarea");
      textArea.value = textToCopy;
      textArea.style.position = "fixed";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        const successful = document.execCommand("copy");
        if (!successful) throw new Error("Fallback: Copy command was unsuccessful");
        window.open(linkToOpen, '_blank');
      } catch (err: any) {
        const message = err?.message || "Fallback: Copy failed";
        console.log(message);
      } finally {
        document.body.removeChild(textArea);
      }
    } else {
      navigator.clipboard.writeText(textToCopy)
        .then(() => window.open(linkToOpen, '_blank'))
        .catch((err: any) => {
          const message = err?.message || "Clipboard copy failed";
          console.log(message);
        });
    }
  };

  // // Show account rotation toast when post has a recently used account
  // useEffect(() => {
  //   posts.forEach(post => {
  //     if (shouldSuggestAccountRotation(post)) {
  //       toast({
  //         title: "Consider rotating accounts",
  //         description: `The account u/${redditAccounts.find(acc => acc.id === post.assignedAccountId)?.username} was used recently. Consider using a different account for this post.`,
  //         duration: 5000,
  //       });
  //     }
  //   });
  //   // eslint-disable-next-line react-hooks/exhaustive-deps
  // }, []);

  const handleLeadStatusUpdate = async (status: LeadStatus, leadId: string) => {
    setSelectedLeadId(leadId);
    try {
      await portalClient.updateLeadStatus({ status, leadId });

      const allLeads = newTabList.filter((lead) => lead.id !== leadId);
      dispatch(setNewTabList(allLeads));
      // toast({
      //   title: "Post saved",
      //   description: "This post has been saved for later.",
      // });
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || "Something went wrong";
      toast({
        title: "Error",
        description: message,
      });
    } finally {
      setSelectedLeadId("");
    }
  };

  if (newTabList?.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <MessageSquare className="h-12 w-12 text-muted-foreground mb-4" />
        <h4 className="text-lg font-medium mb-2">{`Sit back and relax, we are finding relevant leads for you. We will notify you once it’s ready.`}</h4>
      </div>
    );
  }

  return (
    <ScrollArea className="h-[calc(100vh-300px)]">
      <div className="space-y-4 pr-4">
        {newTabList?.map(post => (
          <Card key={post.id} className="overflow-hidden">
            <CardContent className="p-6">
              <div className="flex flex-col space-y-4">
                {/* Header with score and title */}
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-1">
                      <div className={`text-sm font-bold px-2 py-1 rounded-md ${getScoreColor(post.relevancyScore)}`}>
                        {post.relevancyScore}%
                      </div>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className="cursor-pointer">
                              <Brain className="h-4 w-4 text-primary" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="top" className="max-w-[300px]">
                            <MarkdownRenderer data={post.metadata?.chainOfThought || ""} />
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>

                    <Link href={post.metadata?.postUrl || "#"} target="_blank">
                      <h3 className="text-lg font-medium">
                        <HtmlTitleRenderer htmlString={post.title || ""} />
                      </h3>
                    </Link>
                    <div className="flex items-center text-sm text-muted-foreground gap-2 mt-1">
                      {/* <span>{getSubredditName(subredditList, post?.sourceId ?? "")}</span> */}
                      <span>{post.metadata?.subredditPrefixed}</span>
                      <>{console.log("###12", subredditList)}</>
                      <span>•</span>
                      <span>{getFormattedDate(post.postCreatedAt)}</span>
                      <span>•</span>
                      <Link href={post.metadata?.authorUrl || "#"} target="_blank">
                        by {post.author}
                      </Link>
                    </div>
                  </div>
                </div>

                <ScrollArea className="h-40">
                  <HtmlBodyRenderer htmlString={post.metadata?.descriptionHtml || ""} />
                </ScrollArea>

                {/* Last matched info */}
                <div className="text-xs text-muted-foreground">
                  Last matched: {post.createdAt ? getFormattedDate(post.createdAt) : "N/A"}
                </div>

                {/* Snippet and AI suggestion */}
                {post.metadata?.suggestedComment && (
                  <div className="bg-secondary/50 rounded-md p-3">
                    <div className="flex items-center gap-2 text-sm font-medium mb-2">
                      <MessageSquare className="h-4 w-4" />
                      <span className="text-primary">
                        {post?.metadata?.automatedCommentUrl
                          ? "Commented By AI"
                          : post?.metadata?.commentScheduledAt
                            ? "Scheduled by AI"
                            : "Suggested Comment"}
                      </span>
                      {(post?.metadata?.automatedCommentUrl || post?.metadata?.commentScheduledAt) && (
                        <span className="text-xs text-muted-foreground ml-auto">
                          {getFormattedDate(post?.metadata?.commentScheduledAt)}
                        </span>
                      )}

                    </div>
                    <p className="text-sm">
                      <MarkdownRenderer data={post.metadata?.suggestedComment || ""} />
                    </p>

                    <Link
                      href={`https://www.reddit.com/${post.metadata.subredditPrefixed}/about/rules`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="mt-1 text-blue-500 underline"
                    >
                      As per community guidelines
                    </Link>
                  </div>
                )}

                {/* Expanded DM suggestion */}
                {post.metadata?.suggestedDm && (<>
                  {/* {expandedId === post.id && ( */}
                  <div className="bg-secondary/50 rounded-md p-3 border-l-4 border-primary">
                    <div className="flex items-center gap-2 text-sm font-medium mb-2">
                      <MessageSquare className="h-4 w-4 text-primary" />
                      <span className="text-primary">
                        {post?.metadata?.automatedDmSent
                          ? "DM Sent By AI"
                          : post?.metadata?.dmScheduledAt
                            ? "Scheduled by AI"
                            : "Suggested DM"}
                      </span>
                      {(post?.metadata?.automatedDmSent || post?.metadata?.dmScheduledAt) && (
                        <span className="text-xs text-muted-foreground ml-auto">
                          {getFormattedDate(post?.metadata.dmScheduledAt)}
                        </span>
                      )}
                    </div>
                    <MarkdownRenderer data={post.metadata?.suggestedDm || ""} />
                  </div>
                  {/* )} */}
                </>)}

                {/* Reddit account selector with rotation suggestion if needed */}
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2 text-sm">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Account:</span>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div className="relative">
                            {redditAccounts?.[0]?.details?.value?.userName}
                            {/* <RedditAccountSelector
                              accounts={redditAccounts}
                              currentAccountId={post.assignedAccountId || defaultAccountId}
                              onAccountChange={(accountId) => handleAccountChange(post.id, accountId)}
                              postId={post.id}
                            /> */}
                          </div>
                        </TooltipTrigger>
                        {/* <TooltipContent side="top">
                          <p className="text-xs max-w-[200px]">
                            {shouldSuggestAccountRotation(post)
                              ? "This account was used recently. Consider rotating accounts to avoid rate limits."
                              : "Using multiple Reddit accounts helps avoid rate limits and boosts reach."}
                          </p>
                        </TooltipContent> */}
                      </Tooltip>
                    </TooltipProvider>
                  </div>

                  {/* Show last replied time if available */}
                  {/* {post.lastReplied && (
                    <span className="text-xs text-muted-foreground">
                      Last replied: {timeAgo(post.lastReplied)}
                    </span>
                  )} */}
                </div>

                {/* Action buttons */}
                <div className="flex flex-wrap gap-2 mt-2">

                  {post.metadata?.suggestedComment && (
                    <div>
                      <button
                        onClick={() =>
                          copyTextAndOpenLink(
                            post.metadata?.suggestedComment ?? "",
                            post.metadata?.automatedCommentUrl || post.metadata?.postUrl || "#"
                          )
                        }
                        className={`inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input ${post.metadata?.automatedCommentUrl ? "bg-green-600 text-white" : "bg-background hover:bg-accent"} disabled:opacity-50 disabled:pointer-events-none`}
                      >
                        <MessageSquare className="h-4 w-4 mr-2" />
                        {post.metadata?.automatedCommentUrl ? "View Comment" : "Copy & open comment"}
                      </button>
                    </div>
                  )}

                  {post.metadata?.suggestedDm && (
                    <div>
                      <button
                        onClick={() =>
                          copyTextAndOpenLink(
                            post.metadata?.suggestedDm ?? "",
                            post.metadata?.dmUrl ?? "#"
                          )
                        }
                        className={`inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input ${post.metadata?.automatedCommentUrl ? "bg-green-600 text-white" : "bg-background hover:bg-accent"} disabled:opacity-50 disabled:pointer-events-none`}
                      >
                        <Send className="h-4 w-4 mr-2" />
                        {post.metadata?.automatedDmSent ? "View DM" : "Copy & open DM"}
                      </button>
                    </div>
                  )}

                  {(leadStatusFilter === null) && (<>
                    <button
                      onClick={() => handleLeadStatusUpdate(LeadStatus.LEAD, post.id)}
                      className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent"
                      disabled={post.id === selectedLeadId}
                    >
                      <Save className="h-4 w-4 mr-2" />
                      Save as Lead
                    </button>

                    <button
                      onClick={() => handleLeadStatusUpdate(LeadStatus.COMPLETED, post.id)}
                      className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent"
                      disabled={post.id === selectedLeadId}
                    >
                      <ThumbsUp className="h-4 w-4 mr-2" />
                      Mark as Responded
                    </button>

                    <button
                      onClick={() => handleLeadStatusUpdate(LeadStatus.NOT_RELEVANT, post.id)}
                      className="inline-flex items-center justify-center text-sm font-medium h-9 px-4 py-2 rounded-md border border-input bg-background hover:bg-accent"
                    >
                      <X className="h-4 w-4 mr-2" />
                      Skip
                    </button>
                  </>)}

                  {/* {post.metadata?.suggestedDm &&
                    <button
                      onClick={() => toggleExpand(post.id)}
                      className="ml-auto text-sm text-primary hover:underline"
                    >
                      {expandedId === post.id ? 'Hide DM suggestion' : 'Show DM suggestion'}
                    </button>
                  } */}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </ScrollArea>
  );
}