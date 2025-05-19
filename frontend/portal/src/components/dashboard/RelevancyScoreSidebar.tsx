
import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RedditAccount, RedditAccountBadge } from "@/components/reddit-accounts/RedditAccountBadge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Info } from "lucide-react";

interface RelevancyScoreSidebarProps {
  accounts?: RedditAccount[];
  defaultAccountId?: string;
  onDefaultAccountChange?: (accountId: string) => void;
}

export function RelevancyScoreSidebar({
  accounts = [],
  defaultAccountId = "",
  onDefaultAccountChange
}: RelevancyScoreSidebarProps) {
  const defaultAccount = accounts.find(acc => acc.id === defaultAccountId);

  return (
    <Card className="border-primary/10 shadow-md">
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">Relevancy Settings</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">Minimum Score</span>
            <span className="text-sm font-medium">0.75</span>
          </div>
          <Slider defaultValue={[0.75]} max={1} step={0.01} />
          <p className="mt-1 text-xs text-muted-foreground">
            Only show leads with score above this threshold
          </p>
        </div>

        <div className="pt-2 border-t">
          <div className="flex items-start justify-between mb-2">
            <span className="text-sm font-medium">Default Reddit Account</span>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <Info className="h-4 w-4 text-muted-foreground" />
                </TooltipTrigger>
                <TooltipContent side="top">
                  <p className="text-xs max-w-[200px]">Using multiple Reddit accounts helps avoid rate limits and boosts reach.</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>

          {accounts.length > 0 && onDefaultAccountChange ? (
            <Select
              value={defaultAccountId}
              onValueChange={onDefaultAccountChange}
            >
              <SelectTrigger className="w-full">
                <SelectValue>
                  {defaultAccount ? (
                    <div className="flex items-center gap-2">
                      <RedditAccountBadge account={defaultAccount} size="sm" showUsername={false} />
                      <span className="text-xs">u/{defaultAccount.username}</span>
                    </div>
                  ) : (
                    "Select account"
                  )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {accounts.map((account) => {
                  const isActive = !account.status.isBanned &&
                    !account.status.isFlagged &&
                    (!account.status.cooldownMinutes || account.status.cooldownMinutes <= 0);

                  return (
                    <SelectItem
                      key={account.id}
                      value={account.id}
                      disabled={!isActive}
                      className="flex items-center gap-2"
                    >
                      <div className="flex items-center gap-2">
                        <RedditAccountBadge account={account} size="sm" />
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          ) : (
            <p className="text-xs text-muted-foreground">
              <a href="/settings/reddit-accounts" className="text-primary hover:underline">
                Connect Reddit accounts
              </a> to enable automatic replies
            </p>
          )}
          <p className="mt-1 text-xs text-muted-foreground">
            This account will be used for all new leads by default
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
