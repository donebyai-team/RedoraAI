
import React, { useCallback, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
// import {
//   Select,
//   SelectContent,
//   SelectItem,
//   SelectTrigger,
//   SelectValue,
// } from "@/components/ui/select";
// import {
//   RedditAccount,
//   RedditAccountBadge 
// } from "@/components/reddit-accounts/RedditAccountBadge";
// import {
//   Tooltip,
//   TooltipContent,
//   TooltipProvider,
//   TooltipTrigger,
// } from "@/components/ui/tooltip";
// // import { Info } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { RootState } from "@/store/store";
import { useDebounce } from "@doota/ui-core/hooks/useDebounce";
import { setRelevancyScore } from "@/store/Params/ParamsSlice";

export function RelevancyScoreSidebar() {
  const { relevancyScore } = useAppSelector((state: RootState) => state.parems);
  const [relevancy_score, setRelevancy_Score] = useState<number>(relevancyScore);
  const dispatch = useAppDispatch();

  const onChangeCommitted = useCallback((key: string, value: number | string) => {
    if (key === 'relevancy_score') {
      dispatch(setRelevancyScore(value as number));
    }
  }, [dispatch]);

  const debouncedOnChangeCommitted = useDebounce(onChangeCommitted, 700);

  const handleRelevancyChange = (newValue: number[]): void => {
    setRelevancy_Score(newValue[0]);
    debouncedOnChangeCommitted('relevancy_score', newValue[0]);
  }

  return (
    <Card className="border-primary/10 shadow-md">
      <CardHeader className="pb-2">
        <CardTitle className="text-lg">Relevancy Settings</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">Minimum Score</span>
            <span className="text-sm font-medium">{relevancy_score}%</span>
          </div>
          <Slider
            value={[relevancy_score]}
            onValueChange={handleRelevancyChange}
            min={80}
            max={100}
            step={5}
          />
          <p className="mt-1 text-xs text-muted-foreground">
            Only show leads with score above this threshold
          </p>
        </div>

        {/* <div className="pt-2 border-t">
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
        </div> */}
      </CardContent>
    </Card>
  );
}
