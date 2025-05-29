
import { Card, CardContent } from "@/components/ui/card";
import { LeadAnalysis } from "@doota/pb/doota/portal/v1/portal_pb";
import {
  MessageSquare,
  Search,
  Send,
  ArrowUp
} from "lucide-react";

interface PropType {
  counts: LeadAnalysis | undefined,
  loading?: boolean
}

export function SummaryCards({ counts }: PropType) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-secondary/30">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Posts Tracked</p>
              <h3 className="text-2xl font-bold">{counts?.postsTracked ?? "0"}</h3>
            </div>
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <Search className="h-5 w-5 text-primary" />
            </div>
          </div>
          {/* <p className="text-xs text-muted-foreground mt-2">+12% from yesterday</p> */}
        </CardContent>
      </Card>

      <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-purple-500/10">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Relevant Posts</p>
              <h3 className="text-2xl font-bold">{counts?.relevantPostsFound ?? "0"}</h3>
            </div>
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <MessageSquare className="h-5 w-5 text-primary" />
            </div>
          </div>
          {/* <p className="text-xs text-muted-foreground mt-2">+8% from yesterday</p> */}
        </CardContent>
      </Card>

      <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-blue-500/10">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">{(counts?.commentScheduled as number) > 0 ? "Comments Sent/Schedule" : "Comments Sent"}</p>
              <h3 className="text-2xl font-bold">{counts?.commentScheduled ? `${counts.commentSent ?? "0"}/${counts.commentScheduled ?? "0"}` : `${counts?.commentSent ?? "0"}`}</h3>
            </div>
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <Send className="h-5 w-5 text-primary" />
            </div>
          </div>
          {/* <p className="text-xs text-muted-foreground mt-2">+5% from yesterday</p> */}
        </CardContent>
      </Card>

      <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-green-500/10">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">{(counts?.dmScheduled as number) > 0 ? "DM Sent/Schedule" : "DM Sent"}</p>
              <h3 className="text-2xl font-bold">{counts?.dmScheduled ? `${counts.dmSent ?? "0"}/${counts.dmScheduled ?? "0"}` : `${counts?.dmSent ?? "0"}`}</h3>
            </div>
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <ArrowUp className="h-5 w-5 text-primary" />
            </div>
          </div>
          {/* <p className="text-xs text-muted-foreground mt-2">+0.04 from yesterday</p> */}
        </CardContent>
      </Card>
    </div>
  );
}
