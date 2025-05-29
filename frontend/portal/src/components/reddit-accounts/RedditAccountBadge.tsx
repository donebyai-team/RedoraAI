
import React from "react";
import { Check, AlertTriangle, XCircle, User } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";

export interface RedditAccountStatus {
  isActive: boolean;
  hasLowKarma?: boolean;
  isFlagged?: boolean;
  cooldownMinutes?: number;
  isBanned?: boolean;
}

export interface RedditAccount {
  id: string;
  username: string;
  karma: number;
  status: RedditAccountStatus;
  isDefault?: boolean;
}

interface RedditAccountBadgeProps {
  account: RedditAccount;
  size?: "sm" | "md";
  showUsername?: boolean;
}

export function RedditAccountBadge({
  account,
  size = "md",
  showUsername = true,
}: RedditAccountBadgeProps) {
  // Determine status icon and tooltip message
  let StatusIcon = Check;
  let statusColor = "text-green-500";
  let statusBgColor = "bg-green-50";
  let statusMessage = "Active and ready to use";
  let badgeVariant: "default" | "secondary" | "destructive" | "outline" = "default";

  if (account.status.isBanned) {
    StatusIcon = XCircle;
    statusColor = "text-red-500";
    statusBgColor = "bg-red-50";
    statusMessage = "This account has been banned";
    badgeVariant = "destructive";
  } else if (account.status.isFlagged) {
    StatusIcon = XCircle;
    statusColor = "text-red-500";
    statusBgColor = "bg-red-50";
    statusMessage = "This account has been flagged";
    badgeVariant = "destructive";
  } else if (account.status.cooldownMinutes && account.status.cooldownMinutes > 0) {
    StatusIcon = XCircle;
    statusColor = "text-red-500";
    statusBgColor = "bg-red-50";
    statusMessage = `On cooldown for ${account.status.cooldownMinutes} minutes`;
    badgeVariant = "destructive";
  } else if (account.status.hasLowKarma) {
    StatusIcon = AlertTriangle;
    statusColor = "text-amber-500";
    statusBgColor = "bg-amber-50";
    statusMessage = "Low karma (<100) may limit visibility";
    badgeVariant = "secondary";
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className={cn(
            "flex items-center gap-1.5",
            size === "sm" ? "text-xs" : "text-sm"
          )}>
            {showUsername ? (
              <div className="flex items-center gap-1.5">
                <div className="flex items-center justify-center h-5 w-5 rounded-full bg-secondary">
                  <User className="h-3 w-3" />
                </div>
                <span className="font-medium">u/{account.username}</span>
                <Badge variant={badgeVariant} className="px-1.5 py-0.5 h-5 flex items-center gap-1">
                  <StatusIcon className={cn("h-3 w-3", statusColor)} />
                  <span className="text-xs">{account.karma}</span>
                </Badge>
              </div>
            ) : (
              <Badge variant={badgeVariant} className={cn(
                "px-1.5 h-5 flex items-center gap-1",
                statusBgColor
              )}>
                <StatusIcon className={cn("h-3 w-3", statusColor)} />
              </Badge>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent side="top">
          <div className="text-xs">
            <p className="font-medium">u/{account.username}</p>
            <p>{statusMessage}</p>
            <p>Karma: {account.karma}</p>
            {account.isDefault && <p className="text-primary">Default account</p>}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
