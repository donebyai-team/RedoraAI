
import React from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { ChevronDown, Check } from "lucide-react";
import { RedditAccount, RedditAccountBadge } from "./RedditAccountBadge";
import { toast } from "@/components/ui/use-toast";

interface RedditAccountSelectorProps {
  accounts: RedditAccount[];
  currentAccountId: string;
  onAccountChange: (accountId: string) => void;
  postId?: string;
  disabled?: boolean;
}

export function RedditAccountSelector({
  accounts,
  currentAccountId,
  onAccountChange,
  postId,
  disabled = false,
}: RedditAccountSelectorProps) {
  const currentAccount = accounts.find((acc) => acc.id === currentAccountId) || accounts[0];
  
  const handleAccountChange = (accountId: string) => {
    if (accountId === currentAccountId) return;
    
    onAccountChange(accountId);
    
    // Show toast notification
    const newAccount = accounts.find(acc => acc.id === accountId);
    if (newAccount) {
      toast({
        title: "Account changed",
        description: postId 
          ? `Now replying as u/${newAccount.username} for this post`
          : `Default account set to u/${newAccount.username}`,
      });
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild disabled={disabled || accounts.length <= 1}>
        <Button 
          variant="outline" 
          size="sm"
          className="h-8 gap-1 bg-background pr-2 pl-1.5"
        >
          {currentAccount && (
            <>
              <RedditAccountBadge account={currentAccount} size="sm" showUsername={false} />
              <span className="text-xs ml-0.5">u/{currentAccount.username}</span>
            </>
          )}
          <ChevronDown className="h-3.5 w-3.5 ml-0.5 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuLabel>Select Reddit Account</DropdownMenuLabel>
        <DropdownMenuSeparator />
        
        {accounts.map((account) => {
          const isActive = !account.status.isBanned && 
                          !account.status.isFlagged && 
                          (!account.status.cooldownMinutes || account.status.cooldownMinutes <= 0);
          
          return (
            <DropdownMenuItem
              key={account.id}
              disabled={!isActive}
              className="flex justify-between items-center py-2"
              onClick={() => handleAccountChange(account.id)}
            >
              <div className="flex items-center gap-2">
                <RedditAccountBadge account={account} size="sm" />
              </div>
              {account.id === currentAccountId && (
                <Check className="h-4 w-4 text-primary" />
              )}
            </DropdownMenuItem>
          );
        })}
        
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <a href="/settings/reddit-accounts" className="text-xs text-primary cursor-pointer">
            Manage Reddit Accounts
          </a>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
