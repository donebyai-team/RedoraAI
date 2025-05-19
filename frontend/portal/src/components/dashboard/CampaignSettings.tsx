
import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { RedditAccount } from "@/components/reddit-accounts/RedditAccountBadge";
import { RedditAccountSelector } from "@/components/reddit-accounts/RedditAccountSelector";
import { 
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { HelpCircle, RotateCw } from "lucide-react";

interface CampaignSettingsProps {
  accounts: RedditAccount[];
  defaultAccountId: string;
  onDefaultAccountChange: (accountId: string) => void;
}

export function CampaignSettings({ 
  accounts, 
  defaultAccountId, 
  onDefaultAccountChange 
}: CampaignSettingsProps) {
  const [autoRotate, setAutoRotate] = useState(false);
  
  const handleAutoRotateChange = (checked: boolean) => {
    setAutoRotate(checked);
    // In a real app, you would save this setting
  };

  return (
    <Card className="border-primary/10 shadow-md">
      <CardHeader>
        <CardTitle className="text-lg">Reddit Account Settings</CardTitle>
        <CardDescription>
          Configure which Reddit accounts to use for outreach
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="default-account">Default Reddit Account</Label>
              <p className="text-sm text-muted-foreground">
                Used for all posts unless overridden
              </p>
            </div>
            <RedditAccountSelector
              accounts={accounts}
              currentAccountId={defaultAccountId}
              onAccountChange={onDefaultAccountChange}
            />
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5 flex items-center gap-2">
              <Label htmlFor="auto-rotate" className="flex items-center gap-2">
                Auto-rotate Reddit accounts
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="h-4 w-4 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent className="w-80">
                      <p className="text-xs">
                        Using multiple Reddit accounts helps avoid rate limits and boosts reach.
                        When enabled, Redora will automatically rotate through your active accounts
                        when replying to posts.
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <RotateCw className={`h-4 w-4 ${autoRotate ? "text-primary" : "text-muted-foreground"}`} />
              <Switch 
                id="auto-rotate" 
                checked={autoRotate}
                onCheckedChange={handleAutoRotateChange}
              />
            </div>
          </div>
          <p className="text-xs text-muted-foreground pl-0.5">
            Automatically rotates between your active Reddit accounts to avoid rate limits
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
