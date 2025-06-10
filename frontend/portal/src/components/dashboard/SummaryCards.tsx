
import { Card, CardContent } from "@/components/ui/card";
import { LeadAnalysis } from "@doota/pb/doota/portal/v1/portal_pb";
import {
  MessageSquare,
  Search,
  Send,
  // ArrowUp, 
  Pin
} from "lucide-react";
import Link from "next/link";

interface PropType {
  counts: LeadAnalysis | undefined,
  loading?: boolean
}

export function SummaryCards({ counts }: PropType) {
  return (

    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Link href={"/leads"}>
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
      </Link>
      <Link href={"/leads"}>
        <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-purple-500/10">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Relevant Posts</p>
                <h3 className="text-2xl font-bold">{counts?.relevantPostsFound ?? "0"}</h3>
              </div>
              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                <Pin className="h-5 w-5 text-primary" />
              </div>
            </div>
            {/* <p className="text-xs text-muted-foreground mt-2">+8% from yesterday</p> */}
          </CardContent>
        </Card>
      </Link>

      <Link href={"/interactions"}>
        <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-blue-500/10 cursor-pointer">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Comments Sent</p>
                <h3 className="text-2xl font-bold">{counts?.commentSent ?? "0"}</h3>
              </div>
              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                <MessageSquare className="h-5 w-5 text-primary" />
              </div>
            </div>
            {counts?.commentScheduled as number > 0 && (
              <p className="text-xs text-muted-foreground mt-2">
                {counts?.commentScheduled} Scheduled
              </p>
            )}
          </CardContent>
        </Card>
      </Link>

      <Link href={"/interactions"}>
        <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-green-500/10">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">DM Sent</p>
                <h3 className="text-2xl font-bold">{counts?.dmSent ?? "0"}</h3>
              </div>
              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                <Send className="h-5 w-5 text-primary" />
              </div>
            </div>
            {counts?.dmScheduled as number > 0 && (
              <p className="text-xs text-muted-foreground mt-2">
                {counts?.dmScheduled} Scheduled
              </p>
            )}
          </CardContent>
        </Card>
      </Link>
    </div>
  );
}
