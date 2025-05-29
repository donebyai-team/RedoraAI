
import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { 
  AlertCircle, 
  ArrowUpRight, 
  CheckCircle2, 
  Clock, 
  Plus, 
  RefreshCw, 
  Trash2,
  XCircle
} from "lucide-react";
import { 
  Tooltip, 
  TooltipContent, 
  TooltipProvider, 
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface RedditAccount {
  id: string;
  username: string;
  karma: number;
  status: "active" | "disconnected" | "warming_up";
  isDefault: boolean;
}

export function RedditAccountsList() {
  // Mock data - would come from props or API in real implementation
  const [accounts, setAccounts] = useState<RedditAccount[]>([
    { id: "a1", username: "redora_official", karma: 1245, status: "active", isDefault: true },
    { id: "a2", username: "tech_founder", karma: 872, status: "disconnected", isDefault: false },
    { id: "a3", username: "marketing_pro", karma: 3541, status: "warming_up", isDefault: false },
  ]);

  const [autoRotate, setAutoRotate] = useState(false);

  const handleRemoveAccount = (id: string) => {
    setAccounts(accounts.filter(account => account.id !== id));
  };

  const handleToggleDefault = (id: string) => {
    setAccounts(accounts.map(account => ({
      ...account,
      isDefault: account.id === id
    })));
  };

  const getStatusBadge = (status: RedditAccount["status"]) => {
    switch (status) {
      case "active":
        return (
          <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200 flex items-center gap-1">
            <CheckCircle2 className="h-3 w-3" />
            Active
          </Badge>
        );
      case "disconnected":
        return (
          <Badge variant="outline" className="bg-red-50 text-red-700 border-red-200 flex items-center gap-1">
            <XCircle className="h-3 w-3" />
            Disconnected
          </Badge>
        );
      case "warming_up":
        return (
          <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200 flex items-center gap-1">
            <Clock className="h-3 w-3" />
            Warming up
          </Badge>
        );
    }
  };

  return (
    <Card className="border-primary/10 shadow-md">
      <CardHeader>
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              Connected Reddit Accounts
            </CardTitle>
            <CardDescription>
              Add and manage Reddit accounts for this workspace
            </CardDescription>
          </div>
          <Button className="gap-1">
            <Plus className="h-4 w-4" />
            Add Reddit Account
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {accounts.length === 0 ? (
          <div className="text-center py-8 border border-dashed rounded-md">
            <p className="text-muted-foreground mb-4">
              No Reddit accounts connected yet. Add an account to start engaging with leads.
            </p>
            <Button className="gap-1">
              <Plus className="h-4 w-4" />
              Connect Account
            </Button>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-2">
                <Switch 
                  id="auto-rotate" 
                  checked={autoRotate} 
                  onCheckedChange={setAutoRotate}
                />
                <label htmlFor="auto-rotate" className="text-sm cursor-pointer select-none">
                  Auto-rotate Reddit accounts
                </label>
              </div>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                      <AlertCircle className="h-4 w-4" />
                      <span className="sr-only">Info</span>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="max-w-xs">
                      When enabled, Redora will automatically rotate between your active
                      Reddit accounts when posting comments or sending messages.
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            
            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-1">
              {accounts.map((account) => (
                <div 
                  key={account.id}
                  className="flex items-center justify-between p-4 border rounded-md bg-card hover:bg-accent/5"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-secondary/80 flex items-center justify-center">
                      {account.username.charAt(0).toUpperCase()}
                    </div>
                    
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">u/{account.username}</span>
                        <a 
                          href={`https://reddit.com/user/${account.username}`}
                          target="_blank" 
                          rel="noopener noreferrer"
                          className="text-muted-foreground hover:text-primary"
                        >
                          <ArrowUpRight className="h-3 w-3" />
                        </a>
                        {account.isDefault && (
                          <Badge variant="secondary" className="text-xs">
                            Default
                          </Badge>
                        )}
                      </div>
                      
                      <div className="flex items-center gap-3">
                        <span className="text-xs text-muted-foreground">
                          {account.karma.toLocaleString()} karma
                        </span>
                        {getStatusBadge(account.status)}
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    {!account.isDefault && (
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={() => handleToggleDefault(account.id)}
                      >
                        Set as default
                      </Button>
                    )}
                    
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button 
                            variant="ghost" 
                            size="icon"
                            disabled={account.status !== "disconnected"}
                            className="h-8 w-8"
                          >
                            <RefreshCw className="h-4 w-4 text-muted-foreground" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Reconnect account</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button 
                            variant="ghost" 
                            size="icon" 
                            className="h-8 w-8"
                            onClick={() => handleRemoveAccount(account.id)}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Remove account</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
