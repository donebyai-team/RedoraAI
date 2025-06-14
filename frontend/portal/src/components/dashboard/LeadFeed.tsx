
import { useEffect, useRef, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import {
  Brain,
  MessageSquare,
  Save,
  Send,
  ThumbsUp,
  User,
  X,
} from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "@/components/ui/use-toast";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { getFormattedDate } from "@/utils/format";
import Link from "next/link";
import { HtmlBodyRenderer, HtmlTitleRenderer, MarkdownRenderer } from "../Html/HtmlRenderer";
import { LeadStatus } from "@doota/pb/doota/core/v1/core_pb";
import { portalClient } from "@/services/grpc";
import { setLeadList } from "@/store/Lead/leadSlice";
import { loadMoreLeadsProps } from "@/hooks/useLeadListManager";
import { InlineLoader } from "../ui/InlineLoader";

interface LeadFeedProps {
  loadMoreLeads: (props: loadMoreLeadsProps) => Promise<void>;
  isFetchingMore: boolean;
  hasMore: boolean;
}

export function LeadFeed({ loadMoreLeads, hasMore, isFetchingMore }: LeadFeedProps) {

  const [selectedLeadId, setSelectedLeadId] = useState<string>("");
  const { leadList, leadStatusFilter } = useAppSelector((state) => state.lead);
  const { accounts: redditAccounts } = useAppSelector((state) => state.reddit);
  const dispatch = useAppDispatch();
  const scrollRef = useRef<HTMLDivElement | null>(null);

  const getScoreColor = (score: number) => {
    if (score >= 90) return "text-green-500 bg-green-50";
    if (score >= 70) return "text-amber-500 bg-amber-50";
    return "text-red-500 bg-red-50";
  };

  const copyTextAndOpenLink = (textToCopy: string, linkToOpen: string) => {
    const fallbackCopy = () => {
      const textArea = document.createElement("textarea");
      textArea.value = textToCopy;
      textArea.style.position = "fixed";
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        if (document.execCommand("copy")) {
          window.open(linkToOpen, '_blank');
        }
      } finally {
        document.body.removeChild(textArea);
      }
    };

    if (navigator.clipboard?.writeText) {
      navigator.clipboard.writeText(textToCopy)
        .then(() => window.open(linkToOpen, '_blank'))
        .catch(fallbackCopy);
    } else {
      fallbackCopy();
    }
  };

  const handleLeadStatusUpdate = async (status: LeadStatus, leadId: string) => {
    setSelectedLeadId(leadId);
    try {
      await portalClient.updateLeadStatus({ status, leadId });

      const allLeads = leadList.filter((lead) => lead.id !== leadId);
      dispatch(setLeadList(allLeads));
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

  // show post action button (Save as Lead,Mark as Responded,Skip)
  const isShowActionButton = (leadStatusFilter === null || leadStatusFilter === LeadStatus.NEW);

  useEffect(() => {
    const scrollEl = scrollRef.current;
    if (!scrollEl) return;

    const handleScroll = () => {
      if (scrollEl.scrollHeight - scrollEl.scrollTop <= scrollEl.clientHeight + 10) {
        if (hasMore && !isFetchingMore) {
          loadMoreLeads({ isFetchingMore, hasMore });
        }
      }
    };

    scrollEl.addEventListener("scroll", handleScroll);

    return () => {
      scrollEl.removeEventListener("scroll", handleScroll);
    };
  }, [hasMore, isFetchingMore, loadMoreLeads]);

  if (leadList?.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
        <MessageSquare className="h-10 w-10 text-muted-foreground mb-3" />
        <h5 className="text-sm font-normal leading-relaxed text-muted-foreground max-w-md">
          Sit back and relax — we’re finding relevant leads for you. You’ll be notified once it’s ready.
        </h5>
      </div>

    );
  }

  return (
    <ScrollArea viewportRef={scrollRef} className="h-[calc(100vh-300px)]">
      <div className="space-y-4 pr-4">
        {leadList?.map(post => (
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
                          <TooltipContent side="top" className="max-w-[500px] min-w-[400px]">
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
                      <span>•</span>
                      <span>Posted {getFormattedDate(post.postCreatedAt)}</span>
                      <span>•</span>
                      <Link href={post.metadata?.authorUrl || "#"} target="_blank">
                        by {post.author}
                      </Link>
                    </div>
                  </div>
                </div>

                <HtmlBodyRenderer htmlString={post.metadata?.descriptionHtml || ""} />

                {/* Last matched info */}
                <div className="text-xs text-muted-foreground">
                  Matched: {post.createdAt ? getFormattedDate(post.createdAt) : "N/A"}
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
                            ? "Comment Scheduled by AI"
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
                      <Send className="h-4 w-4 text-primary" />
                      <span className="text-primary">
                        {post?.metadata?.automatedDmSent
                          ? "DM Sent By AI"
                          : post?.metadata?.dmScheduledAt
                            ? "DM Scheduled by AI"
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
                            {redditAccounts?.[0]?.details?.value?.userName ? (
                              <a
                                href={`https://reddit.com/user/${redditAccounts[0].details.value.userName}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-blue-600 hover:underline"
                              >
                                {redditAccounts[0].details.value.userName}
                              </a>
                            ) : (
                              '—'
                            )}
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

                  {isShowActionButton ? (<>
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
                  </>) : null}

                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
      <InlineLoader
        isVisible={isFetchingMore}
        size={30}
      />

    </ScrollArea>
  );
}