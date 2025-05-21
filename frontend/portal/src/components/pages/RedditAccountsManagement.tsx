
import React from "react";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { RedditAccountsList } from "@/components/reddit-accounts/RedditAccountsList";
import { Button } from "@/components/ui/button";
import { PlusCircle, Info } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export default function RedditAccountsManagement() {
  return (
    <>
      <DashboardHeader />
      
      <div className="flex-1 overflow-auto">
        <main className="container mx-auto px-4 py-6 md:px-6">
          <div className="space-y-2 mb-6">
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
              Reddit Accounts
            </h1>
            <p className="text-muted-foreground">
              Manage connected Reddit accounts for posting comments and sending messages to leads.
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
            <Card className="border-primary/10 shadow-md">
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Active Accounts</CardTitle>
                <CardDescription>Ready to use</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-primary">3</div>
              </CardContent>
            </Card>
            
            <Card className="border-primary/10 shadow-md">
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Limited Accounts</CardTitle>
                <CardDescription>Low karma or cooldown</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-amber-500">1</div>
              </CardContent>
            </Card>
            
            <Card className="border-primary/10 shadow-md">
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">Flagged</CardTitle>
                <CardDescription>Need attention</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-red-500">1</div>
              </CardContent>
            </Card>
            
            <Card className="border-primary/10 shadow-md bg-gradient-to-br from-background to-secondary/30">
              <CardHeader className="pb-2">
                <CardTitle className="text-lg flex items-center gap-2">
                  Best Practices
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Info className="h-4 w-4 text-muted-foreground" />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="text-xs max-w-xs">Tips for maintaining healthy Reddit accounts</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="text-sm space-y-1 list-disc pl-4">
                  <li>Build karma before outreach</li>
                  <li>Rotate accounts regularly</li>
                  <li>Avoid excessive self-promotion</li>
                </ul>
              </CardContent>
            </Card>
          </div>
          
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Connected Accounts</h2>
            <Button className="gap-2">
              <PlusCircle className="h-4 w-4" />
              <span>Connect Account</span>
            </Button>
          </div>
          
          <RedditAccountsList />
        </main>
      </div>
      
      <DashboardFooter />
    </>
  );
}
